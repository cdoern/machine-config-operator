package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
	"k8s.io/klog/v2"
)

const (
	// the number of times to retry commands that pull data from the network
	NumRetriesNetCommands = 5
	// Default ostreeAuthFile location
	OStreeAuthFile = "/run/ostree/auth.json"
	// Pull secret.  Written by the machine-config-operator
	KubeletAuthFile = "/var/lib/kubelet/config.json"
	// Internal Registry Pull secret + Global Pull secret.  Written by the machine-config-operator.
	ImageRegistryAuthFile = "/etc/mco/internal-registry-pull-secret.json"
)

type LayeredUpdater struct {
}

func (l *LayeredUpdater) ApplyUpdate(newConfig *mcfgv1.MachineConfig) error {
	newEnough, err := isNewEnoughForLayering()
	if err != nil {
		return err
	}
	// If the host isn't new enough to understand the new container model natively, run as a privileged container.
	// See https://github.com/coreos/rpm-ostree/pull/3961 and https://issues.redhat.com/browse/MCO-356
	// This currently will incur a double reboot; see https://github.com/coreos/rpm-ostree/issues/4018
	if !newEnough {
		logSystem("rpm-ostree is not new enough for layering; forcing an update via container")
		if err := inplaceUpdateViaNewContainer(newConfig.Spec.OSImageURL); err != nil {
			return err
		}
	} else if err := rebaseLayered(newConfig.Spec.OSImageURL); err != nil {
		return fmt.Errorf("failed to update OS to %s : %w", newConfig.Spec.OSImageURL, err)
	}

	return nil
}

// InplaceUpdateViaNewContainer runs rpm-ostree ex deploy-via-self
// via a privileged container.  This is needed on firstboot of old
// nodes as well as temporarily for 4.11 -> 4.12 upgrades.
func inplaceUpdateViaNewContainer(target string) error {
	// HACK: Disable selinux enforcement for this because it's not
	// really easily possible to get the correct install_t context
	// here when run from a container image.
	// xref https://issues.redhat.com/browse/MCO-396
	enforceFile := "/sys/fs/selinux/enforce"
	enforcingBuf, err := os.ReadFile(enforceFile)
	var enforcing bool
	if err != nil {
		if os.IsNotExist(err) {
			enforcing = false
		} else {
			return fmt.Errorf("failed to read %s: %w", enforceFile, err)
		}
	} else {
		enforcingStr := string(enforcingBuf)
		v, err := strconv.Atoi(strings.TrimSpace(enforcingStr))
		if err != nil {
			return fmt.Errorf("failed to parse selinux enforcing %v: %w", enforcingBuf, err)
		}
		enforcing = (v == 1)
	}
	if enforcing {
		if err := runCmdSync("setenforce", "0"); err != nil {
			return err
		}
	} else {
		klog.Info("SELinux is not enforcing")
	}

	systemdPodmanArgs := []string{"--unit", "machine-config-daemon-update-rpmostree-via-container", "-p", "EnvironmentFile=-/etc/mco/proxy.env", "--collect", "--wait", "--", "podman"}
	pullArgs := append([]string{}, systemdPodmanArgs...)
	pullArgs = append(pullArgs, "pull", "--authfile", "/var/lib/kubelet/config.json", target)
	err = runCmdSync("systemd-run", pullArgs...)
	if err != nil {
		return err
	}

	runArgs := append([]string{}, systemdPodmanArgs...)
	runArgs = append(runArgs, "run", "--env-file", "/etc/mco/proxy.env", "--privileged", "--pid=host", "--net=host", "--rm", "-v", "/:/run/host", target, "rpm-ostree", "ex", "deploy-from-self", "/run/host")
	err = runCmdSync("systemd-run", runArgs...)
	if err != nil {
		return err
	}
	if enforcing {
		if err := runCmdSync("setenforce", "1"); err != nil {
			return err
		}
	}
	return nil
}

// RebaseLayered rebases system or errors if already rebased
func rebaseLayered(imgURL string) (err error) {
	// Try to re-link the merged pull secrets if they exist, since it could have been populated without a daemon reboot
	useMergedPullSecrets()
	klog.Infof("Executing rebase to %s", imgURL)
	return runRpmOstree("rebase", "--experimental", "ostree-unverified-registry:"+imgURL)
}

// useMergedSecrets gives the rpm-ostree client access to secrets for the internal registry and the global pull
// secret. It does this by symlinking the merged secrets file into /run/ostree. If it fails to find the
// merged secrets, it will use the default pull secret file instead.
func useMergedPullSecrets() error {

	// check if merged secret file exists
	if _, err := os.Stat(ImageRegistryAuthFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			klog.Errorf("Merged secret file does not exist; defaulting to cluster pull secret")
			return linkOstreeAuthFile(KubeletAuthFile)
		}
	}
	// Check that merged secret file is valid JSON
	if file, err := os.ReadFile(ImageRegistryAuthFile); err != nil {
		klog.Errorf("Merged secret file could not be read; defaulting to cluster pull secret %v", err)
		return linkOstreeAuthFile(KubeletAuthFile)
	} else if !json.Valid(file) {
		klog.Errorf("Merged secret file could not be validated; defaulting to cluster pull secret %v", err)
		return linkOstreeAuthFile(KubeletAuthFile)
	}

	// Attempt to link the merged secrets file
	return linkOstreeAuthFile(ImageRegistryAuthFile)
}

// linkOstreeAuthFile gives the rpm-ostree client access to secrets in the file located at `path` by symlinking so that
// rpm-ostree can use those secrets to pull images. This can be called multiple times to overwrite an older link.
func linkOstreeAuthFile(path string) error {
	if _, err := os.Lstat(OStreeAuthFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll("/run/ostree", 0o544); err != nil {
				return err
			}
		}
	} else {
		// Remove older symlink if it exists since it needs to be overwritten
		if err := os.Remove(OStreeAuthFile); err != nil {
			return err
		}
	}

	klog.Infof("Linking ostree authfile to %s", path)
	err := os.Symlink(path, OStreeAuthFile)
	return err
}

// Synchronously invoke rpm-ostree, writing its stdout to our stdout,
// and gathering stderr into a buffer which will be returned in err
// in case of error.
func runRpmOstree(args ...string) error {
	return runCmdSync("rpm-ostree", args...)
}

func (l *LayeredUpdater) RollbackUpdate(img string) error {
	return nil
}

func (l *LayeredUpdater) LiveApply(data []byte) error {
	return nil
}

func NewLayeredUpdater() *LayeredUpdater {
	return &LayeredUpdater{}
}
