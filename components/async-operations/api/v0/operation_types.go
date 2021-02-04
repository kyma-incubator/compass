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

package v0

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperationSpec defines the desired state of Operation
type OperationSpec struct {
	OperationID       string        `json:"operation_id"`
	OperationType     OperationType `json:"operation_type"`
	OperationCategory string        `json:"operation_category"`
	ResourceType      string        `json:"resource_type"`
	ResourceID        string        `json:"resource_id"`
	CorrelationID     string        `json:"correlation_id"`
	WebhookIDs        []string      `json:"webhook_ids,omitempty"`
	RequestData       string        `json:"request_data,omitempty"`
}

// +kubebuilder:validation:Enum=Create;Update;Delete
// +kubebuilder:validation:Required
type OperationType string

// OperationStatus defines the observed state of Operation
type OperationStatus struct {
	Webhooks []Webhook `json:"webhooks,omitempty"`
	Status   Status    `json:"status"`
	Message  string    `json:"message,omitempty"`
}

type Webhook struct {
	WebhookID         string `json:"webhook_id"`
	RetriesCount      int32  `json:"retries_count"`
	WebhookPollURL    string `json:"webhook_poll_url"`
	LastPollTimestamp string `json:"last_poll_timestamp"`
	State             State  `json:"state"`
}

// +kubebuilder:validation:Enum=Success;Failed;Polling
// +kubebuilder:validation:Required
type State string

// +kubebuilder:validation:Enum=Ready;InProgress;Failed
// +kubebuilder:validation:Required
type Status string

// +kubebuilder:object:root=true

// Operation is the Schema for the operations API
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.operation_type`
// +kubebuilder:printcolumn:name="Resource Type",type=string,JSONPath=`.spec.resource_type`
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
