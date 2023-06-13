package operator

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/machine-config-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getOperatorConfiguration(kubeClient client.Client) (*operatorv1.MachineConfiguration, error) {
	conf := &operatorv1.MachineConfiguration{}

	err := kubeClient.Get(context.TODO(),
		types.NamespacedName{
			Name: constants.OperatorConfig,
		}, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
