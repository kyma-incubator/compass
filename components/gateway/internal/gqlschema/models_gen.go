// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlschema

import (
	"fmt"
	"io"
	"strconv"
)

type CredentialData interface {
	IsCredentialData()
}

type CredentialRequestAuth interface {
	IsCredentialRequestAuth()
}

type HealthCheckStatus interface {
	IsHealthCheckStatus()
}

type HealthCheckStatusBase interface {
	IsHealthCheckStatusBase()
}

type API struct {
	Spec       *APISpec    `json:"spec"`
	TargetURL  string      `json:"targetURL"`
	Credential *Credential `json:"credential"`
	Headers    *string     `json:"headers"`
}

type APIInput struct {
	Tbd *string `json:"tbd"`
}

type APISpec struct {
	Type         APISpecType   `json:"type"`
	Data         string        `json:"data"`
	FetchRequest *FetchRequest `json:"fetchRequest"`
}

type Application struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Tenant      string                `json:"tenant"`
	Description *string               `json:"description"`
	Labels      *Labels               `json:"labels"`
	Annotations *string               `json:"annotations"`
	Status      *ApplicationStatus    `json:"status"`
	Webhooks    []*ApplicationWebhook `json:"webhooks"`
	Apis        []*API                `json:"apis"`
	Events      []*Event              `json:"events"`
	Docs        []*Document           `json:"docs"`
}

type ApplicationInput struct {
	Name           string                `json:"name"`
	Description    *string               `json:"description"`
	Labels         *Labels               `json:"labels"`
	Apis           []*APIInput           `json:"apis"`
	Events         []*EventInput         `json:"events"`
	Documentations []*DocumentationInput `json:"documentations"`
}

type ApplicationStatus struct {
	Condition ApplicationStatusCondition `json:"condition"`
	Timestamp string                     `json:"timestamp"`
}

type ApplicationWebhook struct {
	Type       ApplicationWebhookType `json:"type"`
	URL        string                 `json:"url"`
	Credential *Credential            `json:"credential"`
}

type BasicCredentialData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (BasicCredentialData) IsCredentialData() {}

type Credential struct {
	ID          string                `json:"id"`
	Data        CredentialData        `json:"data"`
	RequestAuth CredentialRequestAuth `json:"requestAuth"`
}

type CsrfTokenCredentialRequestAuth struct {
	Token string `json:"token"`
}

func (CsrfTokenCredentialRequestAuth) IsCredentialRequestAuth() {}

type Document struct {
	Data         string        `json:"data"`
	FetchRequest *FetchRequest `json:"fetchRequest"`
}

type DocumentationInput struct {
	Tbd *string `json:"tbd"`
}

type Event struct {
	Spec         *EventSpec    `json:"spec"`
	FetchRequest *FetchRequest `json:"fetchRequest"`
}

type EventInput struct {
	Tbd *string `json:"tbd"`
}

type EventSpec struct {
	Type EventSpecType `json:"type"`
	Data string        `json:"data"`
}

type FetchRequest struct {
	URL        *string             `json:"url"`
	Credential *Credential         `json:"credential"`
	Status     *FetchRequestStatus `json:"status"`
}

type FetchRequestStatus struct {
	Condition FetchRequestStatusCondition `json:"condition"`
	Timestamp string                      `json:"timestamp"`
}

type Label struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}

type ManagementPlaneHealthCheck struct {
	Origin    HealthCheckStatusOrigin    `json:"origin"`
	Condition HealthCheckStatusCondition `json:"condition"`
	Message   *string                    `json:"message"`
	Timestamp string                     `json:"timestamp"`
}

func (ManagementPlaneHealthCheck) IsHealthCheckStatus()     {}
func (ManagementPlaneHealthCheck) IsHealthCheckStatusBase() {}

type OAuthCredentialData struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	URL          string `json:"url"`
}

func (OAuthCredentialData) IsCredentialData() {}

type Runtime struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Tenant           string         `json:"tenant"`
	Labels           Labels         `json:"labels"`
	Annotations      string         `json:"annotations"`
	AgentCredentials *Credential    `json:"agentCredentials"`
	Status           *RuntimeStatus `json:"status"`
}

type RuntimeHealthCheck struct {
	Origin    HealthCheckStatusOrigin          `json:"origin"`
	RuntimeID string                           `json:"runtimeId"`
	Agent     *RuntimeHealthCheckPartialStatus `json:"agent"`
	Events    *RuntimeHealthCheckPartialStatus `json:"events"`
	Gateway   *RuntimeHealthCheckPartialStatus `json:"gateway"`
}

func (RuntimeHealthCheck) IsHealthCheckStatus()     {}
func (RuntimeHealthCheck) IsHealthCheckStatusBase() {}

type RuntimeHealthCheckPartialStatus struct {
	Condition HealthCheckStatusCondition `json:"condition"`
	Message   *string                    `json:"message"`
	Timestamp string                     `json:"timestamp"`
}

type RuntimeStatus struct {
	Condition RuntimeStatusCondition `json:"condition"`
	Timestamp string                 `json:"timestamp"`
}

type APISpecType string

const (
	APISpecTypeOData   APISpecType = "O_DATA"
	APISpecTypeOpenAPI APISpecType = "OPEN_API"
)

var AllAPISpecType = []APISpecType{
	APISpecTypeOData,
	APISpecTypeOpenAPI,
}

func (e APISpecType) IsValid() bool {
	switch e {
	case APISpecTypeOData, APISpecTypeOpenAPI:
		return true
	}
	return false
}

func (e APISpecType) String() string {
	return string(e)
}

func (e *APISpecType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = APISpecType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApiSpecType", str)
	}
	return nil
}

func (e APISpecType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ApplicationStatusCondition string

const (
	ApplicationStatusConditionInitial ApplicationStatusCondition = "INITIAL"
	ApplicationStatusConditionUnknown ApplicationStatusCondition = "UNKNOWN"
	ApplicationStatusConditionReady   ApplicationStatusCondition = "READY"
	ApplicationStatusConditionFailed  ApplicationStatusCondition = "FAILED"
)

var AllApplicationStatusCondition = []ApplicationStatusCondition{
	ApplicationStatusConditionInitial,
	ApplicationStatusConditionUnknown,
	ApplicationStatusConditionReady,
	ApplicationStatusConditionFailed,
}

func (e ApplicationStatusCondition) IsValid() bool {
	switch e {
	case ApplicationStatusConditionInitial, ApplicationStatusConditionUnknown, ApplicationStatusConditionReady, ApplicationStatusConditionFailed:
		return true
	}
	return false
}

func (e ApplicationStatusCondition) String() string {
	return string(e)
}

func (e *ApplicationStatusCondition) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ApplicationStatusCondition(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApplicationStatusCondition", str)
	}
	return nil
}

func (e ApplicationStatusCondition) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ApplicationWebhookType string

const (
	ApplicationWebhookTypeHealthCheck   ApplicationWebhookType = "HEALTH_CHECK"
	ApplicationWebhookTypeConfiguration ApplicationWebhookType = "CONFIGURATION"
)

var AllApplicationWebhookType = []ApplicationWebhookType{
	ApplicationWebhookTypeHealthCheck,
	ApplicationWebhookTypeConfiguration,
}

func (e ApplicationWebhookType) IsValid() bool {
	switch e {
	case ApplicationWebhookTypeHealthCheck, ApplicationWebhookTypeConfiguration:
		return true
	}
	return false
}

func (e ApplicationWebhookType) String() string {
	return string(e)
}

func (e *ApplicationWebhookType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ApplicationWebhookType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApplicationWebhookType", str)
	}
	return nil
}

func (e ApplicationWebhookType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type DocumentType string

const (
	DocumentTypeMarkdown DocumentType = "MARKDOWN"
)

var AllDocumentType = []DocumentType{
	DocumentTypeMarkdown,
}

func (e DocumentType) IsValid() bool {
	switch e {
	case DocumentTypeMarkdown:
		return true
	}
	return false
}

func (e DocumentType) String() string {
	return string(e)
}

func (e *DocumentType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = DocumentType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid DocumentType", str)
	}
	return nil
}

func (e DocumentType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type EventSpecType string

const (
	EventSpecTypeAsyncAPI EventSpecType = "ASYNC_API"
)

var AllEventSpecType = []EventSpecType{
	EventSpecTypeAsyncAPI,
}

func (e EventSpecType) IsValid() bool {
	switch e {
	case EventSpecTypeAsyncAPI:
		return true
	}
	return false
}

func (e EventSpecType) String() string {
	return string(e)
}

func (e *EventSpecType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EventSpecType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EventSpecType", str)
	}
	return nil
}

func (e EventSpecType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type FetchRequestStatusCondition string

const (
	FetchRequestStatusConditionInitial   FetchRequestStatusCondition = "INITIAL"
	FetchRequestStatusConditionSucceeded FetchRequestStatusCondition = "SUCCEEDED"
	FetchRequestStatusConditionFailed    FetchRequestStatusCondition = "FAILED"
)

var AllFetchRequestStatusCondition = []FetchRequestStatusCondition{
	FetchRequestStatusConditionInitial,
	FetchRequestStatusConditionSucceeded,
	FetchRequestStatusConditionFailed,
}

func (e FetchRequestStatusCondition) IsValid() bool {
	switch e {
	case FetchRequestStatusConditionInitial, FetchRequestStatusConditionSucceeded, FetchRequestStatusConditionFailed:
		return true
	}
	return false
}

func (e FetchRequestStatusCondition) String() string {
	return string(e)
}

func (e *FetchRequestStatusCondition) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FetchRequestStatusCondition(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FetchRequestStatusCondition", str)
	}
	return nil
}

func (e FetchRequestStatusCondition) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type HealthCheckStatusCondition string

const (
	HealthCheckStatusConditionSucceeded HealthCheckStatusCondition = "SUCCEEDED"
	HealthCheckStatusConditionFailed    HealthCheckStatusCondition = "FAILED"
)

var AllHealthCheckStatusCondition = []HealthCheckStatusCondition{
	HealthCheckStatusConditionSucceeded,
	HealthCheckStatusConditionFailed,
}

func (e HealthCheckStatusCondition) IsValid() bool {
	switch e {
	case HealthCheckStatusConditionSucceeded, HealthCheckStatusConditionFailed:
		return true
	}
	return false
}

func (e HealthCheckStatusCondition) String() string {
	return string(e)
}

func (e *HealthCheckStatusCondition) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = HealthCheckStatusCondition(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid HealthCheckStatusCondition", str)
	}
	return nil
}

func (e HealthCheckStatusCondition) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type HealthCheckStatusOrigin string

const (
	HealthCheckStatusOriginManagementPlane HealthCheckStatusOrigin = "MANAGEMENT_PLANE"
	HealthCheckStatusOriginRuntime         HealthCheckStatusOrigin = "RUNTIME"
)

var AllHealthCheckStatusOrigin = []HealthCheckStatusOrigin{
	HealthCheckStatusOriginManagementPlane,
	HealthCheckStatusOriginRuntime,
}

func (e HealthCheckStatusOrigin) IsValid() bool {
	switch e {
	case HealthCheckStatusOriginManagementPlane, HealthCheckStatusOriginRuntime:
		return true
	}
	return false
}

func (e HealthCheckStatusOrigin) String() string {
	return string(e)
}

func (e *HealthCheckStatusOrigin) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = HealthCheckStatusOrigin(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid HealthCheckStatusOrigin", str)
	}
	return nil
}

func (e HealthCheckStatusOrigin) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type RuntimeStatusCondition string

const (
	RuntimeStatusConditionInitial RuntimeStatusCondition = "INITIAL"
	RuntimeStatusConditionReady   RuntimeStatusCondition = "READY"
	RuntimeStatusConditionFailed  RuntimeStatusCondition = "FAILED"
)

var AllRuntimeStatusCondition = []RuntimeStatusCondition{
	RuntimeStatusConditionInitial,
	RuntimeStatusConditionReady,
	RuntimeStatusConditionFailed,
}

func (e RuntimeStatusCondition) IsValid() bool {
	switch e {
	case RuntimeStatusConditionInitial, RuntimeStatusConditionReady, RuntimeStatusConditionFailed:
		return true
	}
	return false
}

func (e RuntimeStatusCondition) String() string {
	return string(e)
}

func (e *RuntimeStatusCondition) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RuntimeStatusCondition(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RuntimeStatusCondition", str)
	}
	return nil
}

func (e RuntimeStatusCondition) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
