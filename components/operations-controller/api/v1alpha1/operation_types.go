/*
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
)

// +kubebuilder:validation:Enum=Create;Update;Delete
type OperationType string

// OperationSpec defines the desired state of Operation
type OperationSpec struct {
	OperationID       string        `json:"operation_id"`
	OperationType     OperationType `json:"operation_type"`
	OperationCategory string        `json:"operation_category"`
	ResourceType      string        `json:"resource_type"`
	ResourceID        string        `json:"resource_id"`
	CorrelationID     string        `json:"correlation_id"`
	WebhookIDs        []string      `json:"webhook_ids"`
	RequestObject     string        `json:"request_object"`
}

// +kubebuilder:validation:Enum=Success;Failed;Polling
type State string

const (
	StateSuccess    State = "Success"
	StateFailed     State = "Failed"
	StateInProgress State = "In Progress"
)

type Webhook struct {
	WebhookID         string `json:"webhook_id"`
	RetriesCount      int32  `json:"retries_count"`
	WebhookPollURL    string `json:"webhook_poll_url"`
	LastPollTimestamp string `json:"last_poll_timestamp"`
	State             State  `json:"state"`
}

// +kubebuilder:validation:Enum=Ready;Error
type ConditionType string

const (
	ConditionTypeReady ConditionType = "Ready"
	ConditionTypeError ConditionType = "Error"
)

type Condition struct {
	Type    ConditionType          `json:"type"`
	Status  corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	Message string                 `json:"message,omitempty"`
}

// OperationStatus defines the observed state of Operation
type OperationStatus struct {
	Webhooks           []Webhook   `json:"webhooks,omitempty"`
	Conditions         []Condition `json:"conditions,omitempty"`
	Phase              State       `json:"phase,omitempty"`
	ObservedGeneration int64       `json:"observed_generation,omitempty"`
}

// +kubebuilder:object:root=true

// Operation is the Schema for the operations API
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.operation_type`
// +kubebuilder:printcolumn:name="Resource ID",type=string,JSONPath=`.spec.resource_id`
// +kubebuilder:printcolumn:name="Resource Type",type=string,JSONPath=`.spec.resource_type`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:subresource:status
type Operation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperationSpec   `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OperationList contains a list of Operation
type OperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Operation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Operation{}, &OperationList{})
}
