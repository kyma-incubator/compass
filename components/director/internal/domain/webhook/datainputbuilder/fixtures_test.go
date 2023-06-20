package datainputbuilder_test

import (
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

const (
	ScenarioName                = "scenario-A"
	Tnt                         = "953ac686-5773-4ad0-8eb1-2349e931f852"
	RuntimeID                   = "rt-id"
	RuntimeContextRuntimeID     = "rt-ctx-rt-id"
	RuntimeContextID            = "rt-ctx-id"
	RuntimeContext2ID           = "rt-ctx-id-2"
	ApplicationID               = "04f3568d-3e0c-4f6b-b646-e6979e9d060c"
	Application2ID              = "6f5389cf-4f9e-46b3-9870-624d792d94ad"
	ApplicationTemplateID       = "58963c6f-24f6-4128-a05c-51d5356e7e09"
	ApplicationTemplate2ID      = "b203a000-02af-449b-a8b5-fd0787a9fa4e"
	ApplicationTenantID         = "d456f5a3-9b1f-42d5-aa81-8fde48cddfbd"
	ApplicationTemplateTenantID = "946d2550-5d90-4727-93e7-87c48286a6e7"
	globalSubaccountIDLabelKey  = "global_subaccount_id"
)

var (
	applicationMappings = map[string]*webhook.ApplicationWithLabels{
		ApplicationID: {
			Application: fixApplicationModel(ApplicationID),
			Labels:      fixLabelsMapForApplicationWithLabels(),
			Tenant:      testAppTenantWithLabels,
		},
		Application2ID: {
			Application: fixApplicationModelWithoutTemplate(Application2ID),
			Labels:      fixLabelsMapForApplicationWithLabels(),
			Tenant:      testAppTenantWithLabels,
		},
	}

	applicationTemplateMappings = map[string]*webhook.ApplicationTemplateWithLabels{
		ApplicationTemplateID: {
			ApplicationTemplate: fixApplicationTemplateModel(),
			Labels:              fixLabelsMapForApplicationTemplateWithLabels(),
			Tenant:              testAppTemplateTenantWithLabels,
		},
	}

	runtimeMappings = map[string]*webhook.RuntimeWithLabels{
		RuntimeContextRuntimeID: {
			Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
			Labels:  fixLabelsMapForRuntimeWithLabels(),
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: testRuntimeOwner,
				Labels:                convertLabels(testTenantLabels),
			},
		},
		RuntimeID: {
			Runtime: fixRuntimeModel(RuntimeID),
			Labels:  fixLabelsMapForRuntimeWithLabels(),
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: testRuntimeOwner,
				Labels:                convertLabels(testTenantLabels),
			},
		},
	}

	runtimeContextMappings = map[string]*webhook.RuntimeContextWithLabels{
		RuntimeContextRuntimeID: {
			RuntimeContext: fixRuntimeContextModel(),
			Labels:         fixLabelsMapForRuntimeContextWithLabels(),
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: testRuntimeCtxOwner,
				Labels:                convertLabels(testTenantLabels),
			},
		},
		RuntimeID: {
			RuntimeContext: fixRuntimeContextModelWithRuntimeID(RuntimeID),
			Labels:         fixLabelsMapForRuntimeContextWithLabels(),
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: testRuntimeCtxOwner,
				Labels:                convertLabels(testTenantLabels),
			},
		},
	}
)

func fixApplicationLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"app-label-key": "app-label-value",
	}
}

func fixApplicationLabelsMapWithUnquotableLabels() map[string]interface{} {
	return map[string]interface{}{
		"app-label-key": []string{"app-label-value"},
	}
}

func fixLabelsMapForApplicationWithLabels() map[string]string {
	return map[string]string{
		"app-label-key": "app-label-value",
	}
}

func fixLabelsMapForApplicationWithCompositeLabels() map[string]string {
	return map[string]string{
		"app-label-key": "[\"app-label-value\"]",
	}
}

func fixApplicationTemplateLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"apptemplate-label-key": "apptemplate-label-value",
	}
}

func fixLabelsMapForApplicationTemplateWithLabels() map[string]string {
	return map[string]string{
		"apptemplate-label-key": "apptemplate-label-value",
	}
}

func fixLabelsMapForApplicationTemplateWithSubaccountLabels() map[string]string {
	return map[string]string{
		globalSubaccountIDLabelKey: ApplicationTemplateTenantID,
		"apptemplate-label-key":    "apptemplate-label-value",
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

func fixRuntimeLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"runtime-label-key": "runtime-label-value",
	}
}

func fixLabelsMapForRuntimeWithLabels() map[string]string {
	return map[string]string{
		"runtime-label-key": "runtime-label-value",
	}
}

func fixRuntimeContextLabelsMap() map[string]interface{} {
	return map[string]interface{}{
		"runtime-context-label-key": "runtime-context-label-value",
	}
}

func fixLabelsMapForRuntimeContextWithLabels() map[string]string {
	return map[string]string{
		"runtime-context-label-key": "runtime-context-label-value",
	}
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

func fixApplicationModelWithoutTemplate(applicationID string) *model.Application {
	appModel := fixApplicationModel(applicationID)
	appModel.ApplicationTemplateID = nil
	return appModel
}

func fixApplicationTemplateModel() *model.ApplicationTemplate {
	return &model.ApplicationTemplate{
		ID:                   ApplicationTemplateID,
		Name:                 "application template",
		Description:          str.Ptr("some very detailed description"),
		ApplicationInputJSON: `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`,
	}
}

func unusedAppRepo() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func unusedAppTemplateRepo() *automock.ApplicationTemplateRepository {
	return &automock.ApplicationTemplateRepository{}
}

func unusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func unusedRuntimeCtxRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func unusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func unusedTenantRepo() *automock.TenantRepository {
	return &automock.TenantRepository{}
}

func unusedLabelBuilder() *automock.LabelInputBuilder {
	return &automock.LabelInputBuilder{}
}

func unusedTenantBuilder() *automock.TenantInputBuilder {
	return &automock.TenantInputBuilder{}
}
