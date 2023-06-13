// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/machine-config-operator/pkg/apis/operator.openshift.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperatorConditionApplyConfiguration represents an declarative configuration of the OperatorCondition type for use
// with apply.
type OperatorConditionApplyConfiguration struct {
	Type               *string             `json:"type,omitempty"`
	Status             *v1.ConditionStatus `json:"status,omitempty"`
	LastTransitionTime *metav1.Time        `json:"lastTransitionTime,omitempty"`
	Reason             *string             `json:"reason,omitempty"`
	Message            *string             `json:"message,omitempty"`
}

// OperatorConditionApplyConfiguration constructs an declarative configuration of the OperatorCondition type for use with
// apply.
func OperatorCondition() *OperatorConditionApplyConfiguration {
	return &OperatorConditionApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *OperatorConditionApplyConfiguration) WithType(value string) *OperatorConditionApplyConfiguration {
	b.Type = &value
	return b
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *OperatorConditionApplyConfiguration) WithStatus(value v1.ConditionStatus) *OperatorConditionApplyConfiguration {
	b.Status = &value
	return b
}

// WithLastTransitionTime sets the LastTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastTransitionTime field is set to the value of the last call.
func (b *OperatorConditionApplyConfiguration) WithLastTransitionTime(value metav1.Time) *OperatorConditionApplyConfiguration {
	b.LastTransitionTime = &value
	return b
}

// WithReason sets the Reason field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Reason field is set to the value of the last call.
func (b *OperatorConditionApplyConfiguration) WithReason(value string) *OperatorConditionApplyConfiguration {
	b.Reason = &value
	return b
}

// WithMessage sets the Message field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Message field is set to the value of the last call.
func (b *OperatorConditionApplyConfiguration) WithMessage(value string) *OperatorConditionApplyConfiguration {
	b.Message = &value
	return b
}
