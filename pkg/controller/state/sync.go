package state

import (
	"context"

	v1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type ComponentSyncProgression struct {
	Phase   ComponentSyncProgressionPhase
	Message string
	Error   error
}

type ComponentSyncProgressionPhase string

const (
	FetchingObject       ComponentSyncProgressionPhase = "FetchingObject"
	UpdatingObjectStatus ComponentSyncProgressionPhase = "UpdatingObjectStatus"
	UpdatingObjectconfig ComponentSyncProgressionPhase = "UpdatingObjectStatus"
)

// oc get machine-state
// oc get machine-state/mcc
// oc describe machine-state/mcc
// inside of this, we can have the persistent mcc/mcd/operator/upgrading states.
// though operator and upgrading will look different

// Each of these need an eventhandler ?
// ctrlcommon.NamespacedEventRecorder(eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "machineconfigdaemon", Host: nodeName})),
func (ctrl *Controller) syncAll(syncFuncs []syncFunc, ms string) error {
	for _, syncFunc := range syncFuncs {
		state, err := ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), ms, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if state.Name == syncFunc.name {
			if out := syncFunc.fn(syncFunc.watcher, ms); out != nil {
				return out
			}
		} else {
			if out := syncFunc.fn(syncFunc.watcher, ""); out != nil {
				return out
			}
		}
	}
	return nil
}

func (ctrl *Controller) syncMCC(watcher watch.Interface, ms string) error {
	watchchannel := watcher.ResultChan()
	var msToSync *v1.MachineState
	var err error
	if len(ms) > 0 {
		msToSync, err = ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), ms, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	for event := range watchchannel {
		obj, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}
		//	dn.daemonHealthEvents.Event(dn.node, corev1.EventTypeNormal, "GotNode", "Getting node for MCD")
		msg := obj.Message
		reason := obj.Reason
		annos := obj.Annotations

		ctrl.WriteStatus(v1.ControllerState, msg, reason, annos)

		//there is
		// obj.Message
		// obj.Type (normal)
		// obj.Reason
		// obj.
		// if message is of a type?
		// we need to typify these events somehow
		// the mco currently seems to use the reason to do this
	}
	//node
	//	//getMCP
	//	//applyMCForPool
	//	//sync Ctrl/Pool status
	//render
	//	//getMCP
	//	//applyGenMCToPool
	//	//syncStatus
	//template
	//	//getCConfig
	//	//syncCerts
	//	//applyMCForCConfig
	//	//syncStatus
	//kubelet
	//	// getKubeletCfgs for pool
	//	// get MC assoc. with KubeletCfg
	//	// update annos of KubeletCfg, Update assoc MC.
	return ctrl.WriteSpec(msToSync)
}

func (ctrl *Controller) syncMCD(watcher watch.Interface, ms string) error {
	//syncNode
	// WHAT DO THE OBJECTS in .Event MEAN?!
	//	//get node -. more like got node
	//	//update annos -> state and reason
	//	//update mc if necessary -> triggering update on an MC object
	var msToSync *v1.MachineState
	var err error
	if len(ms) > 0 {
		msToSync, err = ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), ms, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	watchchannel := watcher.ResultChan()
	for event := range watchchannel {
		obj, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}
		//	dn.daemonHealthEvents.Event(dn.node, corev1.EventTypeNormal, "GotNode", "Getting node for MCD")
		msg := obj.Message
		reason := obj.Reason
		annos := obj.Annotations

		ctrl.WriteStatus(v1.DaemonState, msg, reason, annos)

	}
	return ctrl.WriteSpec(msToSync)
}

func (ctrl *Controller) syncMCS(watcher watch.Interface, ms string) error {
	watchchannel := watcher.ResultChan()
	var msToSync *v1.MachineState
	var err error
	if len(ms) > 0 {
		msToSync, err = ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), ms, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	for event := range watchchannel {
		obj, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}
		//	dn.daemonHealthEvents.Event(dn.node, corev1.EventTypeNormal, "GotNode", "Getting node for MCD")
		msg := obj.Message
		reason := obj.Reason
		annos := obj.Annotations

		ctrl.WriteStatus(v1.ServerState, msg, reason, annos)

	}
	return ctrl.WriteSpec(msToSync)
}

func (ctrl *Controller) syncMetrics(watcher watch.Interface, ms string) error {
	// get requests
	// update metrics requested, or register, or degregister
	watchchannel := watcher.ResultChan()
	var msToSync *v1.MachineState
	var err error
	if len(ms) > 0 {
		msToSync, err = ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), ms, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	for event := range watchchannel {
		obj, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}
		//	dn.daemonHealthEvents.Event(dn.node, corev1.EventTypeNormal, "GotNode", "Getting node for MCD")
		msg := obj.Message
		reason := obj.Reason
		annos := obj.Annotations

		ctrl.WriteStatus(v1.UpdatingMetrics, msg, reason, annos)

		// msg == DeregisterMetric
		// reason == MCC_State

	}
	return ctrl.WriteSpec(msToSync)

}

func (ctrl *Controller) syncUpgradingProgression(watcher watch.Interface, ms string) error {
	watchchannel := watcher.ResultChan()
	var msToSync *v1.MachineState
	var err error
	// API item of our kind to sync
	if len(ms) > 0 {
		msToSync, err = ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), ms, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	for event := range watchchannel {
		obj, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}
		//	dn.daemonHealthEvents.Event(dn.node, corev1.EventTypeNormal, "GotNode", "Getting node for MCD")
		msg := obj.Message
		reason := obj.Reason
		annos := obj.Annotations

		ctrl.WriteStatus(v1.UpgradeProgression, msg, reason, annos)

	}
	return ctrl.WriteSpec(msToSync)
}

func (ctrl *Controller) syncOperatorProgression(watcher watch.Interface, ms string) error {
	watchchannel := watcher.ResultChan()
	var msToSync *v1.MachineState
	var err error
	// API item of our kind to sync
	if len(ms) > 0 {
		msToSync, err = ctrl.Mcfgclient.MachineconfigurationV1().MachineStates().Get(context.TODO(), ms, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}
	for event := range watchchannel {
		obj, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}
		//	dn.daemonHealthEvents.Event(dn.node, corev1.EventTypeNormal, "GotNode", "Getting node for MCD")
		msg := obj.Message
		reason := obj.Reason
		annos := obj.Annotations

		ctrl.WriteStatus(v1.OperatorProgression, msg, reason, annos)

	}
	return ctrl.WriteSpec(msToSync)
}
