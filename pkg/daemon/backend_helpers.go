package daemon

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	"k8s.io/klog/v2"
)

const (
	// the number of times to retry commands that pull data from the network
	numRetriesNetCommands = 5
	// Default ostreeAuthFile location
	ostreeAuthFile = "/run/ostree/auth.json"
	// Pull secret.  Written by the machine-config-operator
	kubeletAuthFile = "/var/lib/kubelet/config.json"
	// Internal Registry Pull secret + Global Pull secret.  Written by the machine-config-operator.
	imageRegistryAuthFile = "/etc/mco/internal-registry-pull-secret.json"
)

// imageInspection is a public implementation of
// https://github.com/containers/skopeo/blob/82186b916faa9c8c70cfa922229bafe5ae024dec/cmd/skopeo/inspect.go#L20-L31
type imageInspection struct {
	Name          string `json:",omitempty"`
	Tag           string `json:",omitempty"`
	Digest        digest.Digest
	RepoDigests   []string
	Created       *time.Time
	DockerVersion string
	Labels        map[string]string
	Architecture  string
	Os            string
	Layers        []string
}

// runGetOut executes a command, logging it, and return the stdout output.
func runGetOut(command string, args ...string) ([]byte, error) {
	klog.Infof("Running captured: %s %s", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	rawOut, err := cmd.Output()
	if err != nil {
		errtext := ""
		if e, ok := err.(*exec.ExitError); ok {
			// Trim to max of 256 characters
			errtext = fmt.Sprintf("\n%s", truncate(string(e.Stderr), 256))
		}
		return nil, fmt.Errorf("error running %s %s: %s%s", command, strings.Join(args, " "), err, errtext)
	}
	return rawOut, nil
}

// truncate a string using runes/codepoints as limits.
// This specifically will avoid breaking a UTF-8 value.
func truncate(input string, limit int) string {
	asRunes := []rune(input)
	l := len(asRunes)

	if limit >= l {
		return input
	}

	return fmt.Sprintf("%s [%d more chars]", string(asRunes[:limit]), l-limit)
}
