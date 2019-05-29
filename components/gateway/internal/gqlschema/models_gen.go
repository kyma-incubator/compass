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

type APIDefinition struct {
	ID                    string       `json:"id"`
	Spec                  *APISpec     `json:"spec"`
	TargetURL             string       `json:"targetURL"`
	Credential            *Credential  `json:"credential"`
	AdditionalHeaders     *HttpHeaders `json:"additionalHeaders"`
	AdditionalQueryParams *QueryParams `json:"additionalQueryParams"`
	Version               *Version     `json:"version"`
}

type APIDefinitionInput struct {
	TargetURL         string           `json:"targetURL"`
	Credential        *CredentialInput `json:"credential"`
	Spec              *APISpecInput    `json:"spec"`
	InjectHeaders     *HttpHeaders     `json:"injectHeaders"`
	InjectQueryParams *QueryParams     `json:"injectQueryParams"`
}

type APISpec struct {
	// when fetch request specified, data will be automatically populated
	Data         *string       `json:"data"`
	Format       *SpecFormat   `json:"format"`
	Type         APISpecType   `json:"type"`
	FetchRequest *FetchRequest `json:"fetchRequest"`
}

type APISpecInput struct {
	Data         *string            `json:"data"`
	Type         APISpecType        `json:"type"`
	Format       SpecFormat         `json:"format"`
	FetchRequest *FetchRequestInput `json:"fetchRequest"`
}

type Application struct {
	ID             string                 `json:"id"`
	Tenant         string                 `json:"tenant"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description"`
	Labels         map[string]interface{} `json:"labels"`
	Annotations    map[string]interface{} `json:"annotations"`
	Status         *ApplicationStatus     `json:"status"`
	Webhooks       []*ApplicationWebhook  `json:"webhooks"`
	HealthCheckURL *string                `json:"healthCheckURL"`
	Apis           []*APIDefinition       `json:"apis"`
	EventAPIs      []*EventAPIDefinition  `json:"eventAPIs"`
	Documents      []*Document            `json:"documents"`
}

type ApplicationInput struct {
	Name           string                     `json:"name"`
	Description    *string                    `json:"description"`
	Labels         *Labels                    `json:"labels"`
	Annotations    *Annotations               `json:"annotations"`
	Webhooks       []*ApplicationWebhookInput `json:"webhooks"`
	HealthCheckURL *string                    `json:"healthCheckURL"`
	Apis           []*APIDefinitionInput      `json:"apis"`
	Events         []*EventDefinitionInput    `json:"events"`
	Documents      []*DocumentInput           `json:"documents"`
}

type ApplicationStatus struct {
	Condition ApplicationStatusCondition `json:"condition"`
	Timestamp int                        `json:"timestamp"`
}

type ApplicationWebhook struct {
	ID         string                 `json:"id"`
	Type       ApplicationWebhookType `json:"type"`
	URL        string                 `json:"url"`
	Credential *Credential            `json:"credential"`
}

type ApplicationWebhookInput struct {
	Type       ApplicationWebhookType `json:"type"`
	URL        string                 `json:"url"`
	Credential *CredentialInput       `json:"credential"`
}

type BasicCredentialData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (BasicCredentialData) IsCredentialData() {}

type BasicCredentialDataInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CSRFTokenCredentialRequestAuth struct {
	Token string `json:"token"`
}

type CSRFTokenCredentialRequestAuthInput struct {
	Token string `json:"token"`
}

type Credential struct {
	Data        CredentialData         `json:"data"`
	RequestAuth *CredentialRequestAuth `json:"requestAuth"`
}

type CredentialDataInput struct {
	Basic *BasicCredentialDataInput `json:"basic"`
	Oauth *OAuthCredentialDataInput `json:"oauth"`
}

type CredentialInput struct {
	Data        *CredentialDataInput        `json:"data"`
	RequestAuth *CredentialRequestAuthInput `json:"requestAuth"`
}

type CredentialRequestAuth struct {
	Type CredentialRequestAuthType       `json:"type"`
	Csrf *CSRFTokenCredentialRequestAuth `json:"csrf"`
}

type CredentialRequestAuthInput struct {
	Type CredentialRequestAuthType            `json:"type"`
	Csrf *CSRFTokenCredentialRequestAuthInput `json:"csrf"`
}

type Document struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Format DocumentFormat `json:"format"`
	// for example Service Class, API etc
	Kind         *string       `json:"kind"`
	Data         *string       `json:"data"`
	FetchRequest *FetchRequest `json:"fetchRequest"`
}

type DocumentInput struct {
	Title        string             `json:"title"`
	DisplayName  string             `json:"displayName"`
	Description  string             `json:"description"`
	Format       DocumentFormat     `json:"format"`
	Kind         *string            `json:"kind"`
	Data         *string            `json:"data"`
	FetchRequest *FetchRequestInput `json:"fetchRequest"`
}

type EventAPIDefinition struct {
	ID      string     `json:"id"`
	Spec    *EventSpec `json:"spec"`
	Version *Version   `json:"version"`
}

type EventDefinitionInput struct {
	Spec *EventSpecInput `json:"spec"`
}

type EventSpec struct {
	Data         *string       `json:"data"`
	Type         EventSpecType `json:"type"`
	Format       *SpecFormat   `json:"format"`
	FetchRequest *FetchRequest `json:"fetchRequest"`
}

type EventSpecInput struct {
	Data          *string            `json:"data"`
	EventSpecType EventSpecType      `json:"eventSpecType"`
	FetchRequest  *FetchRequestInput `json:"fetchRequest"`
}

//  Compass performs fetch to validate if request is correct and stores a copy
type FetchRequest struct {
	URL        string              `json:"url"`
	Credential *Credential         `json:"credential"`
	Mode       FetchMode           `json:"mode"`
	Filter     *string             `json:"filter"`
	Status     *FetchRequestStatus `json:"status"`
}

type FetchRequestInput struct {
	URL        string           `json:"url"`
	Credential *CredentialInput `json:"credential"`
	Mode       *FetchMode       `json:"mode"`
	Filter     *string          `json:"filter"`
}

type FetchRequestStatus struct {
	Condition FetchRequestStatusCondition `json:"condition"`
	Timestamp Timestamp                   `json:"timestamp"`
}

type HealthCheck struct {
	Type      HealthCheckType            `json:"type"`
	Condition HealthCheckStatusCondition `json:"condition"`
	Origin    *string                    `json:"origin"`
	Message   *string                    `json:"message"`
	Timestamp Timestamp                  `json:"timestamp"`
}

type LabelFilter struct {
	Label    string          `json:"label"`
	Values   []string        `json:"values"`
	Operator *FilterOperator `json:"operator"`
}

type OAuthCredentialData struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	URL          string `json:"url"`
}

func (OAuthCredentialData) IsCredentialData() {}

type OAuthCredentialDataInput struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	URL          string `json:"url"`
}

type Runtime struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Tenant      Tenant         `json:"tenant"`
	Labels      Labels         `json:"labels"`
	Annotations Annotations    `json:"annotations"`
	Status      *RuntimeStatus `json:"status"`
	// directive for checking auth
	AgentCredential *Credential `json:"agentCredential"`
}

type RuntimeInput struct {
	Name        string       `json:"name"`
	Description *string      `json:"description"`
	Labels      *Labels      `json:"labels"`
	Annotations *Annotations `json:"annotations"`
}

type RuntimeStatus struct {
	Condition RuntimeStatusCondition `json:"condition"`
	Timestamp Timestamp              `json:"timestamp"`
}

type Version struct {
	// for example 4.6
	Value      string `json:"value"`
	Deprecated *bool  `json:"deprecated"`
	// for example 4.5
	DeprecatedSince *string `json:"deprecatedSince"`
	// if true, will be removed in the next version
	ForRemoval *bool `json:"forRemoval"`
}

type APISpecType string

const (
	APISpecTypeOdata   APISpecType = "ODATA"
	APISpecTypeOpenAPI APISpecType = "OPEN_API"
)

var AllAPISpecType = []APISpecType{
	APISpecTypeOdata,
	APISpecTypeOpenAPI,
}

func (e APISpecType) IsValid() bool {
	switch e {
	case APISpecTypeOdata, APISpecTypeOpenAPI:
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
		return fmt.Errorf("%s is not a valid APISpecType", str)
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
	ApplicationWebhookTypeConfigurationChanged ApplicationWebhookType = "CONFIGURATION_CHANGED"
)

var AllApplicationWebhookType = []ApplicationWebhookType{
	ApplicationWebhookTypeConfigurationChanged,
}

func (e ApplicationWebhookType) IsValid() bool {
	switch e {
	case ApplicationWebhookTypeConfigurationChanged:
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

type CredentialRequestAuthType string

const (
	CredentialRequestAuthTypeCsrfToken CredentialRequestAuthType = "CSRF_TOKEN"
)

var AllCredentialRequestAuthType = []CredentialRequestAuthType{
	CredentialRequestAuthTypeCsrfToken,
}

func (e CredentialRequestAuthType) IsValid() bool {
	switch e {
	case CredentialRequestAuthTypeCsrfToken:
		return true
	}
	return false
}

func (e CredentialRequestAuthType) String() string {
	return string(e)
}

func (e *CredentialRequestAuthType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = CredentialRequestAuthType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid CredentialRequestAuthType", str)
	}
	return nil
}

func (e CredentialRequestAuthType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type DocumentFormat string

const (
	DocumentFormatMarkdown DocumentFormat = "MARKDOWN"
)

var AllDocumentFormat = []DocumentFormat{
	DocumentFormatMarkdown,
}

func (e DocumentFormat) IsValid() bool {
	switch e {
	case DocumentFormatMarkdown:
		return true
	}
	return false
}

func (e DocumentFormat) String() string {
	return string(e)
}

func (e *DocumentFormat) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = DocumentFormat(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid DocumentFormat", str)
	}
	return nil
}

func (e DocumentFormat) MarshalGQL(w io.Writer) {
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

type FetchMode string

const (
	FetchModeSingle  FetchMode = "SINGLE"
	FetchModePackage FetchMode = "PACKAGE"
	FetchModeIndex   FetchMode = "INDEX"
)

var AllFetchMode = []FetchMode{
	FetchModeSingle,
	FetchModePackage,
	FetchModeIndex,
}

func (e FetchMode) IsValid() bool {
	switch e {
	case FetchModeSingle, FetchModePackage, FetchModeIndex:
		return true
	}
	return false
}

func (e FetchMode) String() string {
	return string(e)
}

func (e *FetchMode) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FetchMode(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FetchMode", str)
	}
	return nil
}

func (e FetchMode) MarshalGQL(w io.Writer) {
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

type FilterOperator string

const (
	FilterOperatorAll FilterOperator = "ALL"
	FilterOperatorAny FilterOperator = "ANY"
)

var AllFilterOperator = []FilterOperator{
	FilterOperatorAll,
	FilterOperatorAny,
}

func (e FilterOperator) IsValid() bool {
	switch e {
	case FilterOperatorAll, FilterOperatorAny:
		return true
	}
	return false
}

func (e FilterOperator) String() string {
	return string(e)
}

func (e *FilterOperator) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FilterOperator(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FilterOperator", str)
	}
	return nil
}

func (e FilterOperator) MarshalGQL(w io.Writer) {
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

type HealthCheckType string

const (
	HealthCheckTypeManagementPlaneApplicationHealthcheck HealthCheckType = "MANAGEMENT_PLANE_APPLICATION_HEALTHCHECK"
)

var AllHealthCheckType = []HealthCheckType{
	HealthCheckTypeManagementPlaneApplicationHealthcheck,
}

func (e HealthCheckType) IsValid() bool {
	switch e {
	case HealthCheckTypeManagementPlaneApplicationHealthcheck:
		return true
	}
	return false
}

func (e HealthCheckType) String() string {
	return string(e)
}

func (e *HealthCheckType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = HealthCheckType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid HealthCheckType", str)
	}
	return nil
}

func (e HealthCheckType) MarshalGQL(w io.Writer) {
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

type SpecFormat string

const (
	SpecFormatYaml SpecFormat = "YAML"
	SpecFormatJSON SpecFormat = "JSON"
)

var AllSpecFormat = []SpecFormat{
	SpecFormatYaml,
	SpecFormatJSON,
}

func (e SpecFormat) IsValid() bool {
	switch e {
	case SpecFormatYaml, SpecFormatJSON:
		return true
	}
	return false
}

func (e SpecFormat) String() string {
	return string(e)
}

func (e *SpecFormat) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SpecFormat(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SpecFormat", str)
	}
	return nil
}

func (e SpecFormat) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
