package build

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

type NodeBuildController struct {
	*Clients
	*informers

	eventRecorder record.EventRecorder

	nodeHandler func(*corev1.Node)

	syncHandler func(node string) error
	enqueueNode func(*corev1.Node)
}
