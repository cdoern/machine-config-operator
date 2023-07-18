package state

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
)

func (ctrl *Controller) WriteStatus(kind v1.MachineStateType, status, reason string, annos map[string]string) {
	// write status and reason to a struct of some sort
	newCondition := v1.ProgressionCondition{
		Phase:  status,
		Reason: reason,
	}

	// where does this state need to live? ...... mcps are watched for changes.
	// this is because you can apply an mcp. I am unsure if you should be able to apply a
	// machinestate. Maybe if you want to modify some collection behavior.
	// regardless, we will need watchers. But unsure if mcp status is handled this way.
	currState, err := ctrl.Clients.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), string(kind), metav1.GetOptions{})
	if err != nil {
		return
	}

	newState := currState.DeepCopy()
	var state v1.StateProgress
	switch kind {
	case v1.ControllerState:
		state = v1.MCCSync
	case v1.DaemonState:
		state = v1.MCDSync
	case v1.UpgradeProgression:
		state = v1.StateProgress(status) // hmmmm
	case v1.OperatorProgression:
		state = v1.StateProgress(status) // hmmmm
	case v1.UpdatingMetrics:
		state = v1.MetricsSync
	}

	var node *corev1.Node
	if currState.Status.MostRecentState[0].Node != nil {
		node = currState.Status.MostRecentState[0].Node
	} else {
		// this means it is our first time updating this specific
		// MS on this node. We need to find out what node we are on
		node, err = ctrl.Kubeclient.CoreV1().Nodes().Get(context.TODO(), annos["node"], metav1.GetOptions{})
		if err != nil {
			return
		}
	}

	newStateProgress := v1.StateOnNode{
		State: state,
		Node:  node,
		Time:  metav1.Now(),
	}

	// update overall progression
	newState.Status.Progression[newStateProgress] = newCondition

	// update most recent state per node
	for i, s := range newState.Status.MostRecentState {
		if s.Node.Name == node.Name {
			newState.Status.MostRecentState[i] = v1.StateOnNode{
				Node:  node,
				State: state,
			}
		}
	}
	_, err = ctrl.Clients.Mcfgclient.MachineconfigurationV1().MachineStates().Update(context.TODO(), newState, metav1.UpdateOptions{})
	if err != nil {
		return
	}

}

func (ctrl *Controller) WriteSpec(ms *v1.MachineState) error {
	var err error
	if ms != nil {
		_, err = ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Update(context.TODO(), ms, metav1.UpdateOptions{})
	}
	return err
}
