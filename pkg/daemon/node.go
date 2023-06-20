package daemon

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/openshift/machine-config-operator/pkg/daemon/constants"
	corev1 "k8s.io/api/core/v1"
)

// nothing needs to happen in first boot. It is just here.

func (dn *Daemon) loadNodeAnnotations(node *corev1.Node) (*corev1.Node, *string, error) {
	ccAnnotation, err := getNodeAnnotation(node, constants.CurrentMachineConfigAnnotationKey)
	// we need to load the annotations from the file only for the
	// first run.
	// the initial annotations do no need to be set if the node
	// already has annotations.
	if err == nil && ccAnnotation != "" {
		return node, nil, nil
	}

	glog.Infof("No %s annotation on node %s: %v, in cluster bootstrap, loading initial node annotation from %s", constants.CurrentMachineConfigAnnotationKey, node.Name, node.Annotations, constants.InitialNodeAnnotationsFilePath)

	d, err := os.ReadFile(constants.InitialNodeAnnotationsFilePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("failed to read initial annotations from %q: %w", constants.InitialNodeAnnotationsFilePath, err)
	}
	if os.IsNotExist(err) {
		// try currentConfig if, for whatever reason we lost annotations? this is super best effort.
		currentOnDisk, err := dn.getCurrentConfigOnDisk()
		if err == nil {
			glog.Infof("Setting initial node config based on current configuration on disk: %s", currentOnDisk.GetName())
			n, err := dn.nodeWriter.SetAnnotations(map[string]string{constants.CurrentMachineConfigAnnotationKey: currentOnDisk.GetName()})
			return n, nil, err
		}
		return nil, nil, err
	}

	var initial map[string]string
	if err := json.Unmarshal(d, &initial); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal initial annotations: %w", err)
	}

	glog.Infof("Setting initial node config: %s", initial[constants.CurrentMachineConfigAnnotationKey])
	node, err = dn.nodeWriter.SetAnnotations(initial)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set initial annotations: %w", err)
	}

	var pool string
	if val, ok := initial[constants.DropIntoCustomPoolAnnotationKey]; ok {
		pool = val
	}
	return node, &pool, nil
}

// getNodeAnnotation gets the node annotation, unsurprisingly
func getNodeAnnotation(node *corev1.Node, k string) (string, error) {
	return getNodeAnnotationExt(node, k, false)
}

// getNodeAnnotationExt is like getNodeAnnotation, but allows one to customize ENOENT handling
func getNodeAnnotationExt(node *corev1.Node, k string, allowNoent bool) (string, error) {
	v, ok := node.Annotations[k]
	if !ok {
		if !allowNoent {
			return "", fmt.Errorf("%s annotation not found on node '%s'", k, node.Name)
		}
		return "", nil
	}

	return v, nil
}
