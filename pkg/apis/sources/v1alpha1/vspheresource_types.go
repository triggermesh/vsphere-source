/*
Copyright 2019 Triggermesh, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
)

// VSphereSourceSpec defines the desired state of VsphereSource
type VSphereSourceSpec struct {
	VSphereURL         string                   `json:"vsphereUrl"`
	VSphereCredsSecret corev1.SecretKeySelector `json:"vsphereCredsSecret"`
	Sink               *corev1.ObjectReference  `json:"sink,omitempty"`
	ServiceAccountName string                   `json:"serviceAccountName,omitempty"`
}

const (
	VSphereSourceEventType                                              = "vmware.vsphere.event"
	VSphereSourceConditionReady                                         = duckv1alpha1.ConditionReady
	VSphereSourceConditionSinkProvided       duckv1alpha1.ConditionType = "SinkProvided"
	VSphereSourceConditionDeployed           duckv1alpha1.ConditionType = "Deployed"
	VSphereSourceConditionEventTypesProvided duckv1alpha1.ConditionType = "EventTypesProvided"
)

var condSet = duckv1alpha1.NewLivingConditionSet(
	VSphereSourceConditionSinkProvided,
	VSphereSourceConditionDeployed)

// VSphereSourceStatus defines the observed state of VsphereSource
type VSphereSourceStatus struct {
	duckv1alpha1.Status `json:",inline"`

	// SinkURI is the current active sink URI that has been configured for the source.
	// +optional
	SinkURI string `json:"sinkUri,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VSphereSource is the Schema for the vspheresources API
// +k8s:openapi-gen=true
type VSphereSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereSourceSpec   `json:"spec,omitempty"`
	Status VSphereSourceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VSphereSourceList contains a list of VsphereSource
type VSphereSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VSphereSource{}, &VSphereSourceList{})
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *VSphereSourceStatus) InitializeConditions() {
	condSet.Manage(s).InitializeConditions()
}

func (s *VSphereSourceStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return condSet.Manage(s).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (s *VSphereSourceStatus) IsReady() bool {
	return condSet.Manage(s).IsHappy()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *VSphereSourceStatus) MarkSink(uri string) {
	s.SinkURI = uri
	if len(uri) > 0 {
		condSet.Manage(s).MarkTrue(VSphereSourceConditionSinkProvided)
	} else {
		condSet.Manage(s).MarkUnknown(VSphereSourceConditionSinkProvided,
			"SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *VSphereSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	condSet.Manage(s).MarkFalse(VSphereSourceConditionSinkProvided, reason, messageFormat, messageA...)
}

// MarkDeployed sets the condition that the source has been deployed.
func (s *VSphereSourceStatus) MarkDeployed() {
	condSet.Manage(s).MarkTrue(VSphereSourceConditionDeployed)
}

// MarkDeploying sets the condition that the source is deploying.
func (s *VSphereSourceStatus) MarkDeploying(reason, messageFormat string, messageA ...interface{}) {
	condSet.Manage(s).MarkUnknown(VSphereSourceConditionDeployed, reason, messageFormat, messageA...)
}

// MarkNotDeployed sets the condition that the source has not been deployed.
func (s *VSphereSourceStatus) MarkNotDeployed(reason, messageFormat string, messageA ...interface{}) {
	condSet.Manage(s).MarkFalse(VSphereSourceConditionDeployed, reason, messageFormat, messageA...)
}

// MarkEventTypes sets the condition that the source has set its event types.
func (s *VSphereSourceStatus) MarkEventTypes() {
	condSet.Manage(s).MarkTrue(VSphereSourceConditionEventTypesProvided)
}

// MarkNoEventTypes sets the condition that the source does not its event types configured.
func (s *VSphereSourceStatus) MarkNoEventTypes(reason, messageFormat string, messageA ...interface{}) {
	condSet.Manage(s).MarkFalse(VSphereSourceConditionEventTypesProvided, reason, messageFormat, messageA...)
}
