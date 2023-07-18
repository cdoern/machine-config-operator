package state

import (
	"context"
	"fmt"

	mcfgv1listers "github.com/openshift/machine-config-operator/pkg/generated/listers/machineconfiguration.openshift.io/v1"

	mcfgv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
)

func StateControllerForNode(client clientset.Interface, node v1.Node) (*v1.Pod, error) {
	listOptions := metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node.Name}).String(),
	}
	listOptions.LabelSelector = labels.SelectorFromSet(labels.Set{"k8s-app": "machine-config-daemon"}).String()

	podList, err := client.CoreV1().Pods("openshift-machine-config-operator").List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	if len(podList.Items) != 1 {
		if len(podList.Items) == 0 {
			return nil, fmt.Errorf("failed to find MSC for node %s", node.Name)
		}
		return nil, fmt.Errorf("too many (%d) MSC's for node %s", len(podList.Items), node.Name)
	}
	return &podList.Items[0], nil
}

func IsUpgradingProgressionTrue(which mcfgv1.StateProgress, pool mcfgv1.MachineConfigPool, msLister mcfgv1listers.MachineStateLister) (bool, error) {
	ms, err := GetMachineStateForPool(pool, msLister)
	if err != nil {
		return false, err
	}
	for _, stateOnNode := range ms.Status.MostRecentState {
		if stateOnNode.State != which {
			return false, nil
		}
	}
	return true, nil
}

func GetMachineStateForPool(pool mcfgv1.MachineConfigPool, msLister mcfgv1listers.MachineStateLister) (*mcfgv1.MachineState, error) {
	return msLister.Get(fmt.Sprintf("%s-upgrade", pool.Name))
}
