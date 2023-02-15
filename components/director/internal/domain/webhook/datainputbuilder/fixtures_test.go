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
	ScenarioName            = "scenario-A"
	Tnt                     = "953ac686-5773-4ad0-8eb1-2349e931f852"
	RuntimeID               = "rt-id"
	RuntimeContextRuntimeID = "rt-ctx-rt-id"
	RuntimeContextID        = "rt-ctx-id"
	RuntimeContext2ID       = "rt-ctx-id-2"
	FormationTemplateID     = "bda5378d-caa1-4ee4-b8bf-f733e180fbf9"
	FormationID             = "cf7e396b-ee70-4a47-9aff-9fa9bfa466c1"
	testFormationName       = "test-formation"
	ApplicationID           = "04f3568d-3e0c-4f6b-b646-e6979e9d060c"
	Application2ID          = "6f5389cf-4f9e-46b3-9870-624d792d94ad"
	ApplicationTemplateID   = "58963c6f-24f6-4128-a05c-51d5356e7e09"
)

var (
	applicationMappings = map[string]*webhook.ApplicationWithLabels{
		ApplicationID: {
			Application: fixApplicationModel(ApplicationID),
			Labels:      fixApplicationLabelsMap(),
		},
		Application2ID: {
			Application: fixApplicationModelWithoutTemplate(Application2ID),
			Labels:      fixApplicationLabelsMap(),
		},
	}

	applicationTemplateMappings = map[string]*webhook.ApplicationTemplateWithLabels{
		ApplicationTemplateID: {
			ApplicationTemplate: fixApplicationTemplateModel(),
			Labels:              fixApplicationTemplateLabelsMap(),
		},
	}

	runtimeMappings = map[string]*webhook.RuntimeWithLabels{
		RuntimeContextRuntimeID: {
			Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
			Labels:  fixRuntimeLabelsMap(),
		},
		RuntimeID: {
			Runtime: fixRuntimeModel(RuntimeID),
			Labels:  fixRuntimeLabelsMap(),
		},
	}

	runtimeContextMappings = map[string]*webhook.RuntimeContextWithLabels{
		RuntimeContextRuntimeID: {
			RuntimeContext: fixRuntimeContextModel(),
			Labels:         fixRuntimeContextLabelsMap(),
		},
		RuntimeID: {
			RuntimeContext: fixRuntimeContextModelWithRuntimeID(RuntimeID),
			Labels:         fixRuntimeContextLabelsMap(),
		},
	}
)

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

func fixRuntimeContextLabelsMap() map[string]interface{} {
	return map[string]interface{}{
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
