package state

import (
	"context"
	"fmt"
	"reflect"
	"time"

	mcfgv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	v1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	mcfgclientset "github.com/openshift/machine-config-operator/pkg/generated/clientset/versioned"
	mcfginformersv1 "github.com/openshift/machine-config-operator/pkg/generated/informers/externalversions/machineconfiguration.openshift.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	maxRetries = 15
)

type syncFunc struct {
	name    string
	fn      func(watch.Interface, string) error
	watcher watch.Interface
}

// need to establish if this is the sort of thing
// that needs an on/off switch like build controller
// It probably doesn't, as long as the MCC starts this as
// it does other controllers

// it seems that other controllers (besides build) do not have a seprate "non generic"
// controller struct. they just figure out how to do things without data
// but we want bookkeeping, so we will need to probably have a struct that has a new "bookkeeping" style entity
// as well as specific health related structs
type StateControllerConfig struct {
	// maybe this is where we keep the data?
	//Progression v1.

	SubControllers []v1.StateSubController
	// Books v1.HealthControllerBookkeeping
	UpdateDelay time.Duration
}
type StateController interface {
	Run(int, <-chan struct{})
	//UpdateHealth(v1.Health) error
	// UpdateBookkeeping()
}

type informers struct {
	mcpInformer mcfginformersv1.MachineConfigPoolInformer
	msInformer  mcfginformersv1.MachineStateInformer
}
type Clients struct {
	Mcfgclient mcfgclientset.Interface
	Kubeclient clientset.Interface
}

/*
// Controller defines the node controller.
type Controller struct {
	client        mcfgclientset.Interface
	kubeClient    clientset.Interface
	eventRecorder record.EventRecorder

	syncHandler func(node string) error
	enqueueNode func(*corev1.Node)

	nodeLister       corelisterv1.NodeLister
	nodeListerSynced cache.InformerSynced

	queue         workqueue.RateLimitingInterface
	ongoingDrains map[string]time.Time

	cfg Config
}
*/

type Controller struct {
	*Clients
	*informers

	syncHandler         func(key string) error
	enqueueMachineState func(*mcfgv1.MachineState)

	mcpListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	config StateControllerConfig

	listeners []watch.Interface

	// make this simpler, we probably only need an array of sorts
	// but it also has to be easy to pass in (CRD)
	subControllers []v1.StateSubController
	//	upgradeHealthController   UpgradeStateController
	bootstrapHealthController BootstrapStateController
	// operatorHealthController  OperatorStateController
}

func New(
	mcpInformer mcfginformersv1.MachineConfigPoolInformer,
	msInformer mcfginformersv1.MachineStateInformer,
	cfg StateControllerConfig,
	kubeClient clientset.Interface,
	mcfgClient mcfgclientset.Interface,
) *Controller {

	ctrl := &Controller{
		informers: &informers{
			mcpInformer: mcpInformer,
			msInformer:  msInformer,
		},
		subControllers: cfg.SubControllers,
		config: StateControllerConfig{
			UpdateDelay: time.Second * 5,
		},
	}

	// does component matter? because I have been using it for whatever I want.
	ctrl.syncHandler = ctrl.syncStateController
	ctrl.enqueueMachineState = ctrl.enqueueDefault
	mccwatcher, err := ctrl.Clients.Kubeclient.CoreV1().Events("openshift-machine-config-operator").Watch(context.TODO(), metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("source", "mcc-health").String()})
	if err != nil {
		klog.Info("watcher failed to generate")
		mccwatcher = nil
	}
	mcdwatcher, err := ctrl.Clients.Kubeclient.CoreV1().Events("openshift-machine-config-operator").Watch(context.TODO(), metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("source", "mcd-health").String()})
	if err != nil {
		klog.Info("watcher failed to generate")
		mcdwatcher = nil
	}
	mcswatcher, err := ctrl.Clients.Kubeclient.CoreV1().Events("openshift-machine-config-operator").Watch(context.TODO(), metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("source", "mcs-health").String()})
	if err != nil {
		klog.Info("watcher failed to generate")
		mcswatcher = nil
	}
	metricswatcher, err := ctrl.Clients.Kubeclient.CoreV1().Events("openshift-machine-config-operator").Watch(context.TODO(), metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("source", "metrics").String()})
	if err != nil {
		klog.Info("watcher failed to generate")
		metricswatcher = nil
	}
	upgradeWatcher, err := ctrl.Clients.Kubeclient.CoreV1().Events("openshift-machine-config-operator").Watch(context.TODO(), metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("source", "upgrade-health").String()})
	if err != nil {
		klog.Info("watcher failed to generate")
		upgradeWatcher = nil
	}
	operatorWatcher, err := ctrl.Clients.Kubeclient.CoreV1().Events("openshift-machine-config-operator").Watch(context.TODO(), metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("source", "operator-health").String()})
	if err != nil {
		klog.Info("watcher failed to generate")
		upgradeWatcher = nil
	}
	bootstrapWatcher, err := ctrl.Clients.Kubeclient.CoreV1().Events("openshift-machine-config-operator").Watch(context.TODO(), metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("source", "bootstrap").String()})
	if err != nil {
		klog.Info("watcher failed to generate")
		upgradeWatcher = nil
	}
	for _, entry := range cfg.SubControllers {
		switch entry {
		//	case v1.StateSubControllerPool:
		//		ctrl.upgradeHealthController = *newUpgradeStateController(ctrl.config, upgradeWatcher)
		//	case v1.StateSubControllerOperator:
		//		ctrl.operatorHealthController = *newOperatorStateController(ctrl.config, operatorWatcher)
		case v1.StateSubControllerBootstrap:
			ctrl.bootstrapHealthController = *newBootstrapStateController(ctrl.config, bootstrapWatcher)
		}
	}

	ctrl.msInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    ctrl.addMachineState,
		UpdateFunc: ctrl.updateMachineState,
		//DeleteFunc: ctrl.deleteMachineState,
	})

	ctrl.listeners = []watch.Interface{
		mccwatcher, mcdwatcher, mcswatcher, metricswatcher, upgradeWatcher, operatorWatcher,
	}

	return ctrl
}

// we want to enqueue it, basically just add it to queue
// so then our sync handler can just do some stuff with it
// namely, call out syncAll.
// wait. But does this mean we do not need events?

// we might not even need a controller, but we should. Have daemon send events
// and have the state controller be the only place machine states are modified

// enqueueAfter will enqueue a pool after the provided amount of time.
func (ctrl *Controller) enqueueAfter(pool *mcfgv1.MachineState, after time.Duration) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pool)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", pool, err))
		return
	}

	ctrl.queue.AddAfter(key, after)
}

// need to look at the diff between the api updated structs and the events
// the api updated structs are only going to be user input
// while the events are going to be internally updated
// enqueueDefault calls a default enqueue function
func (ctrl *Controller) enqueueDefault(pool *mcfgv1.MachineState) {
	ctrl.enqueueAfter(pool, ctrl.config.UpdateDelay)
}

func (ctrl *Controller) addMachineState(obj interface{}) {
	ms := obj.(*mcfgv1.MachineState).DeepCopy()
	klog.V(4).Infof("Adding MachineConfigPool %s", ms.Name)
	ctrl.enqueueMachineState(ms)
}

func (ctrl *Controller) updateMachineState(old, curr interface{}) {
	currMS := curr.(*mcfgv1.MachineState).DeepCopy()
	oldMS := old.(*mcfgv1.MachineState).DeepCopy()
	if !reflect.DeepEqual(oldMS.Status, currMS.Status) {
		klog.Info("user cannot change MachineState status via the API")
		return
	}
	klog.V(4).Infof("updating MachineConfigPool %s", currMS.Name)
	ctrl.enqueueMachineState(currMS)
}

func (ctrl *Controller) Run(workers int, stopCh <-chan struct{}) {

	/*	for _, subctrl := range ctrl.subControllers {
		switch subctrl {
		case v1.StateSubControllerPool:
			go ctrl.upgradeHealthController.Run(workers)
		case v1.StateSubControllerBootstrap: // this can be the only one if it exists
			go ctrl.bootstrapHealthController.Run(workers, stopCh)
			go func() {
				shutdown := func() {
					ctrl.bootstrapHealthController.Stop()
				}
				for {
					select {
					case <-stopCh:
						shutdown()
						return
						//	case <-ctrl.bootstrapHealthController.Done():
						//		// We got a stop signal from the Config Drift Monitor.
						//shutdown()
						//		return
					}
				}
			}()
		case v1.StateSubControllerOperator:
			go ctrl.operatorHealthController.Run(workers)
		}
	}*/

	if len(ctrl.subControllers) > 0 && ctrl.subControllers[0] == v1.StateSubControllerBootstrap {
		go ctrl.bootstrapHealthController.Run(workers, stopCh)
		go func() {
			shutdown := func() {
				ctrl.bootstrapHealthController.Stop()
			}
			for {
				select {
				case <-stopCh:
					shutdown()
					return
					//	case <-ctrl.bootstrapHealthController.Done():
					//		// We got a stop signal from the Config Drift Monitor.
					//shutdown()
					//		return
				}
			}
		}()
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ctrl.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (ctrl *Controller) worker() {
	for ctrl.processNextWorkItem() {
	}
}

func (ctrl *Controller) processNextWorkItem() bool {
	key, quit := ctrl.queue.Get()
	if quit {
		return false
	}
	defer ctrl.queue.Done(key)

	err := ctrl.syncHandler(key.(string))
	ctrl.handleErr(err, key)

	return true
}

func (ctrl *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		ctrl.queue.Forget(key)
		return
	}

	if ctrl.queue.NumRequeues(key) < maxRetries {
		klog.V(2).Infof("Error syncing state controller %v: %v", key, err)
		ctrl.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	klog.V(2).Infof("Dropping state controller %q out of the queue: %v", key, err)
	ctrl.queue.Forget(key)
	ctrl.queue.AddAfter(key, 1*time.Minute)
}

func (ctrl *Controller) syncStateController(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing machine state %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing machine state %q (%v)", key, time.Since(startTime))
	}()

	var syncFuncs = []syncFunc{
		{string(v1.ControllerState), ctrl.syncMCC, ctrl.listeners[0]},
		{string(v1.DaemonState), ctrl.syncMCD, ctrl.listeners[1]},
		{string(v1.ServerState), ctrl.syncMCS, ctrl.listeners[2]},
		{string(v1.MetricsSync), ctrl.syncMetrics, ctrl.listeners[3]},
		{string(v1.UpgradeProgression), ctrl.syncUpgradingProgression, ctrl.listeners[4]},
		{string(v1.OperatorProgression), ctrl.syncOperatorProgression, ctrl.listeners[5]},
	}
	return ctrl.syncAll(syncFuncs, key)
}
