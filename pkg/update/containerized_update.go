package update

import mcfgv1 "github.com/openshift/api/machineconfiguration/v1"

type ContainerizedUpdater struct{}

func (c *ContainerizedUpdater) ApplyUpdate(newConfig *mcfgv1.MachineConfig) error {
	// somehow double check the bootc nature of the image
	// bootc, err := isBootcEnabledImage
	if err := bootcUpdate(newConfig.Spec.OSImageURL); err != nil {
		return err
	}

	return nil
}

func (c *ContainerizedUpdater) RollbackUpdate(img string) error {
	return nil
}

func (c *ContainerizedUpdater) LiveApply(data []byte) error {
	return nil
}

func NewContainerizedUpdater() *ContainerizedUpdater {
	return &ContainerizedUpdater{}
}

func bootcUpdate(img string) error {
	return runCmdSync("bootc", "update")
	// or runCmdSync("bootc", "switch", osImageURL)
}
