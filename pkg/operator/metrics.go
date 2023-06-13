package operator

import (
	ctrlcommon "github.com/openshift/machine-config-operator/pkg/controller/common"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	DefaultBindAddress = ":8797"
)

// MCO Metrics
var (
	// mcoState is the state of the machine config operator
	// pause, updated, updating, degraded
	MCOState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mco_state",
			Help: "state of a specified pool",
		}, []string{"pool", "state", "reason"})
	// mcoMachineCount is the total number of nodes in the pool
	MCOMachineCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mco_machine_count",
			Help: "total number of machines in a specified pool",
		}, []string{"pool"})
	// mcoUpdatedMachineCount is the updated machines in the pool
	MCOUpdatedMachineCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mco_updated_machine_count",
			Help: "total number of updated machines in specified pool",
		}, []string{"pool"})
	// mcoDegradedMachineCount is the degraded machines in the pool
	MCODegradedMachineCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mco_degraded_machine_count",
			Help: "total number of degraded machines in specified pool",
		}, []string{"pool"})
	// mcoUnavailableMachineCount is the degraded machines in the pool
	MCOUnavailableMachineCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mco_unavailable_machine_count",
			Help: "total number of unavailable machines in specified pool",
		}, []string{"pool"})
)

func RegisterMCOMetrics() error {
	return ctrlcommon.RegisterMetrics([]prometheus.Collector{MCOState, MCOMachineCount, MCOUpdatedMachineCount, MCODegradedMachineCount, MCOUnavailableMachineCount})
}
