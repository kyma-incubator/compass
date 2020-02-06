// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql

import (
	"fmt"
	"io"
	"strconv"
)

type CredentialData interface {
	IsCredentialData()
}

type OneTimeToken interface {
	IsOneTimeToken()
}

// Every query that implements pagination returns object that implements Pageable interface.
// To specify page details, query specify two parameters: `first` and `after`.
// `first` specify page size, `after` is a cursor for the next page. When requesting first page, set `after` to empty value.
// For requesting next page, set `after` to `pageInfo.endCursor` returned from previous query.
type Pageable interface {
	IsPageable()
}

type APIDefinitionInput struct {
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	TargetURL   string        `json:"targetURL"`
	Group       *string       `json:"group"`
	Spec        *APISpecInput `json:"spec"`
	Version     *VersionInput `json:"version"`
	DefaultAuth *AuthInput    `json:"defaultAuth"`
}

type APIDefinitionPage struct {
	Data       []*APIDefinition `json:"data"`
	PageInfo   *PageInfo        `json:"pageInfo"`
	TotalCount int              `json:"totalCount"`
}

func (APIDefinitionPage) IsPageable() {}

// Deprecated
type APIRuntimeAuth struct {
	RuntimeID string `json:"runtimeID"`
	Auth      *Auth  `json:"auth"`
}

type APISpecInput struct {
	Data         *CLOB              `json:"data"`
	Type         APISpecType        `json:"type"`
	Format       SpecFormat         `json:"format"`
	FetchRequest *FetchRequestInput `json:"fetchRequest"`
}

type ApplicationEventingConfiguration struct {
	DefaultURL string `json:"defaultURL"`
}

type ApplicationFromTemplateInput struct {
	TemplateName string                `json:"templateName"`
	Values       []*TemplateValueInput `json:"values"`
}

type ApplicationPage struct {
	Data       []*Application `json:"data"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

func (ApplicationPage) IsPageable() {}

type ApplicationRegisterInput struct {
	Name                string                          `json:"name"`
	ProviderName        *string                         `json:"providerName"`
	Description         *string                         `json:"description"`
	Labels              *Labels                         `json:"labels"`
	Webhooks            []*WebhookInput                 `json:"webhooks"`
	HealthCheckURL      *string                         `json:"healthCheckURL"`
	APIDefinitions      []*APIDefinitionInput           `json:"apiDefinitions"`
	EventDefinitions    []*EventDefinitionInput         `json:"eventDefinitions"`
	Documents           []*DocumentInput                `json:"documents"`
	Packages            []*PackageDefinitionCreateInput `json:"packages"`
	IntegrationSystemID *string                         `json:"integrationSystemID"`
}

type ApplicationStatus struct {
	Condition ApplicationStatusCondition `json:"condition"`
	Timestamp Timestamp                  `json:"timestamp"`
}

type ApplicationTemplate struct {
	ID               string                         `json:"id"`
	Name             string                         `json:"name"`
	Description      *string                        `json:"description"`
	ApplicationInput string                         `json:"applicationInput"`
	Placeholders     []*PlaceholderDefinition       `json:"placeholders"`
	AccessLevel      ApplicationTemplateAccessLevel `json:"accessLevel"`
}

type ApplicationTemplateInput struct {
	Name             string                         `json:"name"`
	Description      *string                        `json:"description"`
	ApplicationInput *ApplicationRegisterInput      `json:"applicationInput"`
	Placeholders     []*PlaceholderDefinitionInput  `json:"placeholders"`
	AccessLevel      ApplicationTemplateAccessLevel `json:"accessLevel"`
}

type ApplicationTemplatePage struct {
	Data       []*ApplicationTemplate `json:"data"`
	PageInfo   *PageInfo              `json:"pageInfo"`
	TotalCount int                    `json:"totalCount"`
}

func (ApplicationTemplatePage) IsPageable() {}

type ApplicationUpdateInput struct {
	Name                string  `json:"name"`
	ProviderName        *string `json:"providerName"`
	Description         *string `json:"description"`
	HealthCheckURL      *string `json:"healthCheckURL"`
	IntegrationSystemID *string `json:"integrationSystemID"`
}

type Auth struct {
	Credential            CredentialData         `json:"credential"`
	AdditionalHeaders     *HttpHeaders           `json:"additionalHeaders"`
	AdditionalQueryParams *QueryParams           `json:"additionalQueryParams"`
	RequestAuth           *CredentialRequestAuth `json:"requestAuth"`
}

type AuthInput struct {
	Credential            *CredentialDataInput        `json:"credential"`
	AdditionalHeaders     *HttpHeaders                `json:"additionalHeaders"`
	AdditionalQueryParams *QueryParams                `json:"additionalQueryParams"`
	RequestAuth           *CredentialRequestAuthInput `json:"requestAuth"`
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
	TokenEndpointURL      string         `json:"tokenEndpointURL"`
	Credential            CredentialData `json:"credential"`
	AdditionalHeaders     *HttpHeaders   `json:"additionalHeaders"`
	AdditionalQueryParams *QueryParams   `json:"additionalQueryParams"`
}

type CSRFTokenCredentialRequestAuthInput struct {
	TokenEndpointURL      string               `json:"tokenEndpointURL"`
	Credential            *CredentialDataInput `json:"credential"`
	AdditionalHeaders     *HttpHeaders         `json:"additionalHeaders"`
	AdditionalQueryParams *QueryParams         `json:"additionalQueryParams"`
}

type CredentialDataInput struct {
	Basic *BasicCredentialDataInput `json:"basic"`
	Oauth *OAuthCredentialDataInput `json:"oauth"`
}

type CredentialRequestAuth struct {
	Csrf *CSRFTokenCredentialRequestAuth `json:"csrf"`
}

type CredentialRequestAuthInput struct {
	Csrf *CSRFTokenCredentialRequestAuthInput `json:"csrf"`
}

type DocumentInput struct {
	Title        string             `json:"title"`
	DisplayName  string             `json:"displayName"`
	Description  string             `json:"description"`
	Format       DocumentFormat     `json:"format"`
	Kind         *string            `json:"kind"`
	Data         *CLOB              `json:"data"`
	FetchRequest *FetchRequestInput `json:"fetchRequest"`
}

type DocumentPage struct {
	Data       []*Document `json:"data"`
	PageInfo   *PageInfo   `json:"pageInfo"`
	TotalCount int         `json:"totalCount"`
}

func (DocumentPage) IsPageable() {}

type EventDefinition struct {
	ID            string  `json:"id"`
	ApplicationID string  `json:"applicationID"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	// group allows you to find the same API but in different version
	Group   *string    `json:"group"`
	Spec    *EventSpec `json:"spec"`
	Version *Version   `json:"version"`
}

type EventDefinitionInput struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Spec        *EventSpecInput `json:"spec"`
	Group       *string         `json:"group"`
	Version     *VersionInput   `json:"version"`
}

type EventDefinitionPage struct {
	Data       []*EventDefinition `json:"data"`
	PageInfo   *PageInfo          `json:"pageInfo"`
	TotalCount int                `json:"totalCount"`
}

func (EventDefinitionPage) IsPageable() {}

type EventSpecInput struct {
	Data         *CLOB              `json:"data"`
	Type         EventSpecType      `json:"type"`
	Format       SpecFormat         `json:"format"`
	FetchRequest *FetchRequestInput `json:"fetchRequest"`
}

// Compass performs fetch to validate if request is correct and stores a copy
type FetchRequest struct {
	URL    string              `json:"url"`
	Auth   *Auth               `json:"auth"`
	Mode   FetchMode           `json:"mode"`
	Filter *string             `json:"filter"`
	Status *FetchRequestStatus `json:"status"`
}

type FetchRequestInput struct {
	URL    string     `json:"url"`
	Auth   *AuthInput `json:"auth"`
	Mode   *FetchMode `json:"mode"`
	Filter *string    `json:"filter"`
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

type HealthCheckPage struct {
	Data       []*HealthCheck `json:"data"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

func (HealthCheckPage) IsPageable() {}

type InstanceAPIAuth struct {
	ID string `json:"id"`
	// Context of InstanceAPIAuth - such as Runtime ID, namespace
	Context *interface{} `json:"context"`
	// It may be empty if status is PENDING.
	// Populated with `package.defaultAuth` value if `package.defaultAuth` is defined. If not, Compass notifies Application/Integration System about the Auth request.
	Auth   *Auth            `json:"auth"`
	Status *InstanceAPIAuth `json:"status"`
}

type InstanceAPIAuthInput struct {
	PackageID string `json:"packageID"`
	// Context of InstanceAPIAuth - such as Runtime ID, namespace
	Context *interface{} `json:"context"`
	// JSON validated against package.authRequestJSONSchema
	InputParams *interface{} `json:"inputParams"`
}

type InstanceAPIAuthStatus struct {
	Condition *InstanceAPIAuthStatusCondition `json:"condition"`
	Timestamp Timestamp                       `json:"timestamp"`
	Message   *string                         `json:"message"`
	// PendingNotification
	// NotificationSent
	Reason *string `json:"reason"`
}

type IntegrationSystemInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type IntegrationSystemPage struct {
	Data       []*IntegrationSystem `json:"data"`
	PageInfo   *PageInfo            `json:"pageInfo"`
	TotalCount int                  `json:"totalCount"`
}

func (IntegrationSystemPage) IsPageable() {}

type Label struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type LabelDefinition struct {
	Key    string      `json:"key"`
	Schema *JSONSchema `json:"schema"`
}

type LabelDefinitionInput struct {
	Key    string      `json:"key"`
	Schema *JSONSchema `json:"schema"`
}

type LabelFilter struct {
	// Label key. If query for the filter is not provided, returns every object with given label key regardless of its value.
	Key string `json:"key"`
	// Optional SQL/JSON Path expression. If query is not provided, returns every object with given label key regardless of its value.
	// Currently only a limited subset of expressions is supported.
	Query *string `json:"query"`
}

type LabelInput struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type OAuthCredentialData struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	// URL for getting access token
	URL string `json:"url"`
}

func (OAuthCredentialData) IsCredentialData() {}

type OAuthCredentialDataInput struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	URL          string `json:"url"`
}

type PackageDefinition struct {
	ID                    string             `json:"id"`
	Name                  string             `json:"name"`
	Description           *string            `json:"description"`
	AuthRequestJSONSchema *JSONSchema        `json:"authRequestJSONSchema"`
	Auth                  *InstanceAPIAuth   `json:"auth"`
	Auths                 []*InstanceAPIAuth `json:"auths"`
	// When defined, all Auth requests fallback to defaultAuth.
	DefaultAuth      *Auth                `json:"defaultAuth"`
	APIDefinitions   *APIDefinitionPage   `json:"apiDefinitions"`
	EventDefinitions *EventDefinitionPage `json:"eventDefinitions"`
	Documents        *DocumentPage        `json:"documents"`
	APIDefinition    *APIDefinition       `json:"apiDefinition"`
	EventDefinition  *EventDefinition     `json:"eventDefinition"`
	Document         *Document            `json:"document"`
}

type PackageDefinitionCreateInput struct {
	Name                  string                  `json:"name"`
	Description           string                  `json:"description"`
	AuthRequestJSONSchema *JSONSchema             `json:"authRequestJSONSchema"`
	DefaultAuth           *AuthInput              `json:"defaultAuth"`
	APIDefinitions        []*APIDefinitionInput   `json:"apiDefinitions"`
	EventDefinitions      []*EventDefinitionInput `json:"eventDefinitions"`
	Documents             []*DocumentInput        `json:"documents"`
}

type PackageDefinitionUpdateInput struct {
	Name                  string      `json:"name"`
	Description           string      `json:"description"`
	AuthRequestJSONSchema *JSONSchema `json:"authRequestJSONSchema"`
	// While updating defaultAuth, existing InstanceApiAuths are NOT updated.
	DefaultAuth *AuthInput `json:"defaultAuth"`
}

type PageInfo struct {
	StartCursor PageCursor `json:"startCursor"`
	EndCursor   PageCursor `json:"endCursor"`
	HasNextPage bool       `json:"hasNextPage"`
}

type PlaceholderDefinition struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type PlaceholderDefinitionInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type RuntimeEventingConfiguration struct {
	DefaultURL string `json:"defaultURL"`
}

type RuntimeInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Labels      *Labels `json:"labels"`
}

type RuntimeMetadata struct {
	CreationTimestamp Timestamp `json:"creationTimestamp"`
}

type RuntimePage struct {
	Data       []*Runtime `json:"data"`
	PageInfo   *PageInfo  `json:"pageInfo"`
	TotalCount int        `json:"totalCount"`
}

func (RuntimePage) IsPageable() {}

type RuntimeStatus struct {
	Condition RuntimeStatusCondition `json:"condition"`
	Timestamp Timestamp              `json:"timestamp"`
}

type SystemAuth struct {
	ID   string `json:"id"`
	Auth *Auth  `json:"auth"`
}

type TemplateValueInput struct {
	Placeholder string `json:"placeholder"`
	Value       string `json:"value"`
}

type Tenant struct {
	ID         string  `json:"id"`
	InternalID string  `json:"internalID"`
	Name       *string `json:"name"`
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

type VersionInput struct {
	Value           string  `json:"value"`
	Deprecated      *bool   `json:"deprecated"`
	DeprecatedSince *string `json:"deprecatedSince"`
	ForRemoval      *bool   `json:"forRemoval"`
}

type Viewer struct {
	ID   string     `json:"id"`
	Type ViewerType `json:"type"`
}

type Webhook struct {
	ID            string                 `json:"id"`
	ApplicationID string                 `json:"applicationID"`
	Type          ApplicationWebhookType `json:"type"`
	URL           string                 `json:"url"`
	Auth          *Auth                  `json:"auth"`
}

type WebhookInput struct {
	Type ApplicationWebhookType `json:"type"`
	URL  string                 `json:"url"`
	Auth *AuthInput             `json:"auth"`
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

type ApplicationTemplateAccessLevel string

const (
	ApplicationTemplateAccessLevelGlobal ApplicationTemplateAccessLevel = "GLOBAL"
)

var AllApplicationTemplateAccessLevel = []ApplicationTemplateAccessLevel{
	ApplicationTemplateAccessLevelGlobal,
}

func (e ApplicationTemplateAccessLevel) IsValid() bool {
	switch e {
	case ApplicationTemplateAccessLevelGlobal:
		return true
	}
	return false
}

func (e ApplicationTemplateAccessLevel) String() string {
	return string(e)
}

func (e *ApplicationTemplateAccessLevel) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ApplicationTemplateAccessLevel(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApplicationTemplateAccessLevel", str)
	}
	return nil
}

func (e ApplicationTemplateAccessLevel) MarshalGQL(w io.Writer) {
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

type InstanceAPIAuthStatusCondition string

const (
	// When creating or deleting new one
	InstanceAPIAuthStatusConditionPending   InstanceAPIAuthStatusCondition = "PENDING"
	InstanceAPIAuthStatusConditionSucceeded InstanceAPIAuthStatusCondition = "SUCCEEDED"
	InstanceAPIAuthStatusConditionFailed    InstanceAPIAuthStatusCondition = "FAILED"
)

var AllInstanceAPIAuthStatusCondition = []InstanceAPIAuthStatusCondition{
	InstanceAPIAuthStatusConditionPending,
	InstanceAPIAuthStatusConditionSucceeded,
	InstanceAPIAuthStatusConditionFailed,
}

func (e InstanceAPIAuthStatusCondition) IsValid() bool {
	switch e {
	case InstanceAPIAuthStatusConditionPending, InstanceAPIAuthStatusConditionSucceeded, InstanceAPIAuthStatusConditionFailed:
		return true
	}
	return false
}

func (e InstanceAPIAuthStatusCondition) String() string {
	return string(e)
}

func (e *InstanceAPIAuthStatusCondition) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = InstanceAPIAuthStatusCondition(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid InstanceAPIAuthStatusCondition", str)
	}
	return nil
}

func (e InstanceAPIAuthStatusCondition) MarshalGQL(w io.Writer) {
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
	SpecFormatXML  SpecFormat = "XML"
)

var AllSpecFormat = []SpecFormat{
	SpecFormatYaml,
	SpecFormatJSON,
	SpecFormatXML,
}

func (e SpecFormat) IsValid() bool {
	switch e {
	case SpecFormatYaml, SpecFormatJSON, SpecFormatXML:
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

type ViewerType string

const (
	ViewerTypeRuntime           ViewerType = "RUNTIME"
	ViewerTypeApplication       ViewerType = "APPLICATION"
	ViewerTypeIntegrationSystem ViewerType = "INTEGRATION_SYSTEM"
	ViewerTypeUser              ViewerType = "USER"
)

var AllViewerType = []ViewerType{
	ViewerTypeRuntime,
	ViewerTypeApplication,
	ViewerTypeIntegrationSystem,
	ViewerTypeUser,
}

func (e ViewerType) IsValid() bool {
	switch e {
	case ViewerTypeRuntime, ViewerTypeApplication, ViewerTypeIntegrationSystem, ViewerTypeUser:
		return true
	}
	return false
}

func (e ViewerType) String() string {
	return string(e)
}

func (e *ViewerType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ViewerType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ViewerType", str)
	}
	return nil
}

func (e ViewerType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
