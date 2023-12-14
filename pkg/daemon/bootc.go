package daemon

import (
	rpmostreeclient "github.com/coreos/rpmostree-client-go/pkg/client"
	"k8s.io/klog/v2"
)

// RpmOstreeClient provides all RpmOstree related methods in one structure.
// This structure implements DeploymentClient
//
// TODO(runcom): make this private to pkg/daemon!!!
type BootcClient struct {
	client rpmostreeclient.Client
}

func (r *BootcClient) BootcUpdate(imgURL string) (err error) {
	// Try to re-link the merged pull secrets if they exist, since it could have been populated without a daemon reboot
	//useMergedPullSecrets()
	klog.Infof("Executing rebase to %s", imgURL)
	return runBootc("upgrade")
	// return runBootc("switch", imgURL)
	// not sure if we need to switch or if we will be using the same image?
}

func runBootc(args ...string) error {
	return runCmdSync("bootc", args...)
}
