package formation_test

import (
	"context"
	"encoding/json"
	"time"

	databuilderautomock "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

var (
	tenantID          = uuid.New()
	externalTenantID  = uuid.New()
	nilFormationModel *model.Formation

	modelFormation = model.Formation{
		ID:                  FormationID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
	graphqlFormation = graphql.Formation{
		ID:                  FormationID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
	testScenario = "test-scenario"

	formationTemplate = model.FormationTemplate{
		ID:           FormationTemplateID,
		Name:         "formation-template",
		RuntimeTypes: []string{runtimeType},
	}
	runtimeLblFilters = []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("runtimeType", fmt.Sprintf(`$[*] ? (@ == "%s")`, runtimeType))}

	TestConfigValueRawJSON = json.RawMessage(`{"configKey":"configValue"}`)
	TestConfigValueStr     = "{\"configKey\":\"configValue\"}"
)

const (
	TargetTenantID                   = "targetTenantID"
	ScenarioName                     = "scenario-A"
	ScenarioName2                    = "scenario-B"
	ErrMsg                           = "some error"
	Tnt                              = "953ac686-5773-4ad0-8eb1-2349e931f852"
	TargetTenant                     = "targetTenant"
	ExternalTnt                      = "external-tnt"
	TenantID2                        = "18271026-3998-4391-be58-b783a09fcca8"
	TargetTenantID2                  = "targetTenantID2"
	WebhookID                        = "b5a62a7d-6805-43f9-a3be-370d2d125f0f"
	Webhook2ID                       = "b9a62a7d-6805-43f9-a3be-370d2d125f0f"
	RuntimeID                        = "rt-id"
	WebhookForRuntimeContextID       = "5202f196-46d7-4d1e-be50-434dd9fcd157"
	AppTenantMappingWebhookIDForApp1 = "b91e7d97-65ed-4b72-a225-4a3b484c27e1"
	AppTenantMappingWebhookIDForApp2 = "df7e9387-7bdf-46bb-b0c2-de5ec9a40a21"
	RuntimeContextRuntimeID          = "rt-ctx-rt-id"
	RuntimeContextID                 = "rt-ctx-id"
	RuntimeContext2ID                = "rt-ctx-id-2"
	FormationTemplateID              = "bda5378d-caa1-4ee4-b8bf-f733e180fbf9"
	FormationID                      = "cf7e396b-ee70-4a47-9aff-9fa9bfa466c1"
	testFormationName                = "test-formation"
	secondTestFormationName          = "second-formation"
	testFormationTemplateName        = "test-formation-template"
	ApplicationID                    = "04f3568d-3e0c-4f6b-b646-e6979e9d060c"
	Application2ID                   = "6f5389cf-4f9e-46b3-9870-624d792d94ad"
	ApplicationTemplateID            = "58963c6f-24f6-4128-a05c-51d5356e7e09"
	runtimeType                      = "runtimeType"
	applicationType                  = "applicationType"

	// Formation Assignment constants
	FormationAssignmentID          = "FormationAssignmentID"
	FormationAssignmentFormationID = "FormationAssignmentFormationID"
	FormationAssignmentTenantID    = "FormationAssignmentTenantID"
	FormationAssignmentSource      = "FormationAssignmentSource"
	FormationAssignmentSourceType  = "FormationAssignmentSourceType"
	FormationAssignmentTarget      = "FormationAssignmentTarget"
	FormationAssignmentTargetType  = "FormationAssignmentTargetType"
	FormationAssignmentState       = "FormationAssignmentState"
)

func unusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func unusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func unusedASAService() *automock.AutomaticFormationAssignmentService {
	return &automock.AutomaticFormationAssignmentService{}
}

func unusedASARepo() *automock.AutomaticFormationAssignmentRepository {
	return &automock.AutomaticFormationAssignmentRepository{}
}

func unusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func unusedRuntimeContextRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func unusedDataInputBuilder() *databuilderautomock.DataInputBuilder {
	return &databuilderautomock.DataInputBuilder{}
}

func expectEmptySliceRuntimeContextRepo() *automock.RuntimeContextRepository {
	repo := &automock.RuntimeContextRepository{}
	repo.On("ListByIDs", mock.Anything, Tnt, []string{}).Return(nil, nil).Once()
	return repo
}

func unusedApplicationRepo() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func unusedWebhookRepository() *automock.WebhookRepository {
	return &automock.WebhookRepository{}
}

func unusedAppTemplateRepository() *automock.ApplicationTemplateRepository {
	return &automock.ApplicationTemplateRepository{}
}

func unusedWebhookConverter() *automock.WebhookConverter {
	return &automock.WebhookConverter{}
}

func unusedWebhookClient() *automock.WebhookClient {
	return &automock.WebhookClient{}
}

func unusedLabelDefService() *automock.LabelDefService {
	return &automock.LabelDefService{}
}

func unusedUUIDService() *automock.UuidService {
	return &automock.UuidService{}
}

func unusedConverter() *automock.Converter {
	return &automock.Converter{}
}

func unusedService() *automock.Service {
	return &automock.Service{}
}

func unusedFormationRepo() *automock.FormationRepository {
	return &automock.FormationRepository{}
}

func unusedNotificationsService() *automock.NotificationsService {
	return &automock.NotificationsService{}
}

func unusedFormationAssignmentService() *automock.FormationAssignmentService {
	return &automock.FormationAssignmentService{}
}

func noActionNotificationsService() *automock.NotificationsService {
	notificationSvc := &automock.NotificationsService{}
	notificationSvc.On("GenerateNotifications", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	return notificationSvc
}

func unusedFormationTemplateRepo() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func fixCtxWithTenant() context.Context {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String(), externalTenantID.String())

	return ctx
}

func fixModel(scenarioName string) model.AutomaticScenarioAssignment {
	return fixModelWithScenarioName(scenarioName)
}

func fixModelWithScenarioName(scenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   scenario,
		Tenant:         tenantID.String(),
		TargetTenantID: TargetTenantID,
	}
}

func fixError() error {
	return errors.New(ErrMsg)
}

func mockScenarioDefServiceThatReturns(scenarios []string) *automock.LabelDefService {
	mockScenarioDefSvc := &automock.LabelDefService{}
	mockScenarioDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(scenarios, nil)
	return mockScenarioDefSvc
}

func fixUUID() string {
	return FormationID
}

func fixColumns() []string {
	return []string{"id", "tenant_id", "formation_template_id", "name"}
}

func fixScenariosLabelDefinition(tenantID string, schema interface{}) model.LabelDefinition {
	return model.LabelDefinition{
		Key:     model.ScenariosKey,
		Tenant:  tenantID,
		Schema:  &schema,
		Version: 1,
	}
}

func fixFormationTemplateModel() *model.FormationTemplate {
	return &model.FormationTemplate{
		ID:                     FormationTemplateID,
		Name:                   "formation-tmpl-name",
		ApplicationTypes:       []string{"appType1", "appType2"},
		RuntimeTypes:           []string{"runtimeTypes"},
		RuntimeTypeDisplayName: "runtimeDisplayName",
		RuntimeArtifactKind:    model.RuntimeArtifactKindEnvironmentInstance,
	}
}

func fixApplicationModelWithoutTemplate(applicationID string) *model.Application {
	appModel := fixApplicationModel(applicationID)
	appModel.ApplicationTemplateID = nil
	return appModel
}

func fixApplicationModel(applicationID string) *model.Application {
	return &model.Application{
		ProviderName:          str.Ptr("application-provider"),
		ApplicationTemplateID: str.Ptr(ApplicationTemplateID),
		Name:                  "application-name",
		Description:           str.Ptr("detailed application description"),
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: time.Time{},
		},
		HealthCheckURL:      str.Ptr("localhost/healthz"),
		BaseURL:             str.Ptr("base_url"),
		OrdLabels:           json.RawMessage("[]"),
		CorrelationIDs:      json.RawMessage("[]"),
		SystemStatus:        str.Ptr("reachable"),
		DocumentationLabels: json.RawMessage("[]"),
		BaseEntity: &model.BaseEntity{
			ID:        applicationID,
			Ready:     true,
			Error:     nil,
			CreatedAt: &time.Time{},
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixApplicationLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"app-label-key": "app-label-value",
	}
}

func fixApplicationTemplateLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"apptemplate-label-key": "apptemplate-label-value",
	}
}

func fixRuntimeLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"runtime-label-key": "runtime-label-value",
	}
}

func fixRuntimeContextLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"runtime-context-label-key": "runtime-context-label-value",
	}
}

func fixConfigurationChangedWebhookModel(webhookID, objectID string, objectType model.WebhookReferenceObjectType) *model.Webhook {
	return &model.Webhook{
		ID:         webhookID,
		ObjectID:   objectID,
		ObjectType: objectType,
		Type:       model.WebhookTypeConfigurationChanged,
	}
}

func fixApplicationTenantMappingWebhookModel(webhookID, appID string) *model.Webhook {
	return &model.Webhook{
		ID:         webhookID,
		ObjectID:   appID,
		ObjectType: model.ApplicationWebhookReference,
		Type:       model.WebhookTypeApplicationTenantMapping,
	}
}

func fixRuntimeWebhookGQLModel(webhookID, runtimeID string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:        webhookID,
		RuntimeID: str.Ptr(runtimeID),
		Type:      graphql.WebhookTypeConfigurationChanged,
	}
}

func fixApplicationWebhookGQLModel(webhookID, appID string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:            webhookID,
		ApplicationID: str.Ptr(appID),
		Type:          graphql.WebhookTypeConfigurationChanged,
	}
}

func fixApplicationTenantMappingWebhookGQLModel(webhookID, appID string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:        webhookID,
		RuntimeID: str.Ptr(appID),
		Type:      graphql.WebhookTypeApplicationTenantMapping,
	}
}

func fixApplicationTemplateModel() *model.ApplicationTemplate {
	return &model.ApplicationTemplate{
		ID:                   ApplicationTemplateID,
		Name:                 "application template",
		Description:          str.Ptr("some very detailed description"),
		ApplicationInputJSON: `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`,
	}
}

func fixRuntimeModel(runtimeID string) *model.Runtime {
	return &model.Runtime{
		ID:                runtimeID,
		Name:              "runtime name",
		Description:       str.Ptr("some description"),
		CreationTimestamp: time.Time{},
	}
}

func fixRuntimeWithLabels(runtimeID string) *webhook.RuntimeWithLabels {
	return &webhook.RuntimeWithLabels{
		Runtime: fixRuntimeModel(runtimeID),
		Labels:  fixRuntimeLabelsMap(),
	}
}

func fixRuntimeContextModel() *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        RuntimeContextID,
		RuntimeID: RuntimeContextRuntimeID,
		Key:       "some-key",
		Value:     "some-value",
	}
}

func fixRuntimeContextModelWithRuntimeID(rtID string) *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        RuntimeContext2ID,
		RuntimeID: rtID,
		Key:       "some-key",
		Value:     "some-value",
	}
}

func fixFormationModel() *model.Formation {
	return &model.Formation{
		ID:                  FormationID,
		TenantID:            Tnt,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
}

func fixFormationEntity() *formation.Entity {
	return &formation.Entity{
		ID:                  FormationID,
		TenantID:            Tnt,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
}

func fixGqlFormation() *graphql.Formation {
	return &graphql.Formation{
		ID:                  FormationID,
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
	}
}

func fixGqlFormationAssignment(configValue *string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:         FormationAssignmentID,
		Source:     FormationAssignmentSource,
		SourceType: FormationAssignmentSourceType,
		Target:     FormationAssignmentTarget,
		TargetType: FormationAssignmentTargetType,
		State:      FormationAssignmentState,
		Value:      configValue,
	}
}

func fixGqlFormationAssignmentWithSuffix(configValue *string, suffix string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:         FormationAssignmentID + suffix,
		Source:     FormationAssignmentSource + suffix,
		SourceType: graphql.FormationAssignmentType(FormationAssignmentSourceType + suffix),
		Target:     FormationAssignmentTarget + suffix,
		TargetType: graphql.FormationAssignmentType(FormationAssignmentTargetType + suffix),
		State:      FormationAssignmentState + suffix,
		Value:      configValue,
	}
}

func fixFormationAssignmentModel(configValue json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          FormationAssignmentID,
		FormationID: FormationAssignmentFormationID,
		TenantID:    FormationAssignmentTenantID,
		Source:      FormationAssignmentSource,
		SourceType:  FormationAssignmentSourceType,
		Target:      FormationAssignmentTarget,
		TargetType:  FormationAssignmentTargetType,
		State:       FormationAssignmentState,
		Value:       configValue,
	}
}

func fixFormationAssignmentModelWithSuffix(configValue json.RawMessage, suffix string) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          FormationAssignmentID + suffix,
		FormationID: FormationAssignmentFormationID + suffix,
		TenantID:    FormationAssignmentTenantID + suffix,
		Source:      FormationAssignmentSource + suffix,
		SourceType:  model.FormationAssignmentType(FormationAssignmentSourceType + suffix),
		Target:      FormationAssignmentTarget + suffix,
		TargetType:  model.FormationAssignmentType(FormationAssignmentTargetType + suffix),
		State:       FormationAssignmentState + suffix,
		Value:       configValue,
	}
}

func fixFormationAssignmentPage(fas []*model.FormationAssignment) *model.FormationAssignmentPage {
	return &model.FormationAssignmentPage{
		Data: fas,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(fas),
	}
}

func fixGQLFormationAssignmentPage(gqlFAS []*graphql.FormationAssignment) *graphql.FormationAssignmentPage {
	return &graphql.FormationAssignmentPage{
		Data: gqlFAS,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(gqlFAS),
	}
}
