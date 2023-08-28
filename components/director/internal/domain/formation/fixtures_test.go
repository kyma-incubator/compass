package formation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	databuilderautomock "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	// Tenant IDs
	TntInternalID = "953ac686-5773-4ad0-8eb1-2349e931f852"
	TntExternalID = "ada4241d-caa1-4ee4-b8bf-f733e180fbf9"
	TntCustomerID = "ede0241d-caa1-4ee4-b8bf-f733e180fbf9"

	// Automatic Scenario Assignment(ASA) constants
	TargetTenantID  = "targetTenantID-ASA"
	TargetTenantID2 = "targetTenantID2-ASA"
	TenantID2       = "18271026-3998-4391-be58-b783a09fcca8" // used as tenant where the ASA "lives"
	ScenarioName    = "scenario-A"
	ScenarioName2   = "scenario-B"

	// Entity constants
	ApplicationID           = "04f3568d-3e0c-4f6b-b646-e6979e9d060c"
	Application2ID          = "6f5389cf-4f9e-46b3-9870-624d792d94ad"
	ApplicationTemplateID   = "58963c6f-24f6-4128-a05c-51d5356e7e09"
	ApplicationTemplate2ID  = "88963c6f-24f6-4128-a05c-51d5356e7e09"
	RuntimeID               = "rt-id"
	RuntimeContextRuntimeID = "rt-ctx-rt-id"
	RuntimeContextID        = "rt-ctx-id"
	RuntimeContext2ID       = "rt-ctx-id-2"
	FormationID             = "cf7e396b-ee70-4a47-9aff-9fa9bfa466c1"

	// Webhook IDs
	WebhookID  = "b5a62a7d-6805-43f9-a3be-370d2d125f0f"
	Webhook2ID = "b9a62a7d-6805-43f9-a3be-370d2d125f0f"
	Webhook3ID = "aaa62a7d-6805-43f9-a3be-370d2d125f0f"
	Webhook4ID = "43fa5d0b-b037-478d-919a-2f0431feedd4"

	TntParentID                      = "ede0241d-caa1-4ee4-b8bf-f733e180fbf9"
	WebhookForRuntimeContextID       = "5202f196-46d7-4d1e-be50-434dd9fcd157"
	AppTenantMappingWebhookIDForApp1 = "b91e7d97-65ed-4b72-a225-4a3b484c27e1"
	AppTenantMappingWebhookIDForApp2 = "df7e9387-7bdf-46bb-b0c2-de5ec9a40a21"
	FormationLifecycleWebhookID      = "517e0235-0d74-4166-a47c-5a577022d468"

	// Formation constants
	testFormationName       = "test-formation-name"
	testFormationState      = string(model.InitialFormationState)
	testFormationEmptyError = "{}"
	secondTestFormationName = "second-formation"
	TargetTenant            = "targetTenant" // used as "assigning tenant" in formation scenarios/flows

	// Formation Template constants
	FormationTemplateID       = "bda5378d-caa1-4ee4-b8bf-f733e180fbf9"
	testFormationTemplateName = "test-formation-template-name"

	// Formation Assignment constants
	FormationAssignmentID          = "FormationAssignmentID"
	FormationAssignmentFormationID = "FormationAssignmentFormationID"
	FormationAssignmentTenantID    = "FormationAssignmentTenantID"
	FormationAssignmentSource      = "FormationAssignmentSource"
	FormationAssignmentSourceType  = "FormationAssignmentSourceType"
	FormationAssignmentTarget      = "FormationAssignmentTarget"
	FormationAssignmentTargetType  = "FormationAssignmentTargetType"
	FormationAssignmentState       = "FormationAssignmentState"

	// Other constants
	ErrMsg          = "some error"
	runtimeType     = "runtimeType"
	applicationType = "applicationType"
	testProvider    = "Compass"
)

var (
	tenantID               = uuid.New()
	externalTenantID       = uuid.New()
	nilFormationModel      *model.Formation
	runtimeTypeDisplayName = str.Ptr("display name")

	testErr = errors.New("Test error")

	CustomerTenantContextPath = &webhook.CustomerTenantContext{
		CustomerID: TntCustomerID,
		AccountID:  nil,
		Path:       str.Ptr(TntExternalID),
	}

	CustomerTenantContextAccount = fixCustomerTenantContext(TntCustomerID, TntExternalID)

	formationModelWithoutError = fixFormationModelWithoutError()
	modelFormation             = model.Formation{
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

	subscriptionRuntimeArtifactKind = model.RuntimeArtifactKindSubscription
	formationTemplate               = model.FormationTemplate{
		ID:                     FormationTemplateID,
		RuntimeArtifactKind:    &subscriptionRuntimeArtifactKind,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		Name:                   testFormationTemplateName,
		RuntimeTypes:           []string{runtimeType},
	}
	runtimeLblFilters = []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("runtimeType", fmt.Sprintf(`$[*] ? (@ == "%s")`, runtimeType))}

	TestConfigValueRawJSON = json.RawMessage(
		`{"configKey":"configValue"}`,
	)
	TestConfigValueStr = "{\"configKey\":\"configValue\"}"

	emptyFormationAssignment = &webhook.FormationAssignment{Value: "\"\""}

	// Formation assignment notification variables
	runtimeCtxNotificationWithAppTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext: &webhook.RuntimeContextWithLabels{
				RuntimeContext: fixRuntimeContextModel(),
				Labels:         fixRuntimeContextLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	runtimeCtxNotificationWithoutAppTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: nil,
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext: &webhook.RuntimeContextWithLabels{
				RuntimeContext: fixRuntimeContextModel(),
				Labels:         fixRuntimeContextLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	appNotificationWithRtmCtxAndTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext: &webhook.RuntimeContextWithLabels{
				RuntimeContext: fixRuntimeContextModel(),
				Labels:         fixRuntimeContextLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	appNotificationWithRtmCtxWithoutTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationWebhookGQLModel(Webhook2ID, Application2ID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: nil,
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext: &webhook.RuntimeContextWithLabels{
				RuntimeContext: fixRuntimeContextModel(),
				Labels:         fixRuntimeContextLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	runtimeNotificationWithAppTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext:    nil,
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	runtimeNotificationWithoutAppTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: nil,
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext:    nil,
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	applicationsMapping = map[string]*webhook.ApplicationWithLabels{
		ApplicationID: {
			Application: fixApplicationModel(ApplicationID),
			Labels:      fixApplicationLabelsMap(),
		},
		Application2ID: {
			Application: fixApplicationModelWithoutTemplate(Application2ID),
			Labels:      fixApplicationLabelsMap(),
		},
	}
	applicationsMapping2 = map[string]*webhook.ApplicationWithLabels{
		Application2ID: {
			Application: fixApplicationModelWithoutTemplate(Application2ID),
			Labels:      fixApplicationLabelsMap(),
		},
	}

	applicationsMappingWithApplicationTemplate = map[string]*webhook.ApplicationWithLabels{
		Application2ID: {
			Application: fixApplicationModelWithTemplateID(Application2ID, ApplicationTemplate2ID),
			Labels:      fixApplicationLabelsMap(),
		},
	}

	emptyApplicationTemplateMappings = map[string]*webhook.ApplicationTemplateWithLabels{}

	applicationTemplateMappings = map[string]*webhook.ApplicationTemplateWithLabels{
		ApplicationTemplateID: {
			ApplicationTemplate: fixApplicationTemplateModel(),
			Labels:              fixApplicationTemplateLabelsMap(),
		},
	}

	applicationTemplateMappings2 = map[string]*webhook.ApplicationTemplateWithLabels{
		ApplicationTemplateID: {
			ApplicationTemplate: fixApplicationTemplateModel(),
			Labels:              fixApplicationTemplateLabelsMap(),
		},
		ApplicationTemplate2ID: {
			ApplicationTemplate: fixApplicationTemplateModelWithID(ApplicationTemplate2ID),
			Labels:              fixApplicationTemplateLabelsMap(),
		},
	}

	runtimeWithLabels             = fixRuntimeWithLabels(RuntimeID)
	runtimeWithRtmCtxWithLabels   = fixRuntimeWithLabels(RuntimeContextRuntimeID)
	emptyRuntimeContextWithLabels *webhook.RuntimeContextWithLabels
	emptyAppTemplateWithLabels    *webhook.ApplicationTemplateWithLabels

	runtimeCtxWithLabels = &webhook.RuntimeContextWithLabels{
		RuntimeContext: fixRuntimeContextModel(),
		Labels:         fixRuntimeContextLabelsMap(),
	}
	runtimeCtx2WithLabels = &webhook.RuntimeContextWithLabels{
		RuntimeContext: fixRuntimeContextModelWithRuntimeID(RuntimeID),
		Labels:         fixRuntimeContextLabelsMap(),
	}

	runtimesMapping = map[string]*webhook.RuntimeWithLabels{
		RuntimeID:               runtimeWithLabels,
		RuntimeContextRuntimeID: runtimeWithRtmCtxWithLabels,
	}

	runtimeContextsMapping = map[string]*webhook.RuntimeContextWithLabels{
		RuntimeContextRuntimeID: runtimeCtxWithLabels,
		RuntimeID:               runtimeCtx2WithLabels,
	}

	runtimeContextsMapping2 = map[string]*webhook.RuntimeContextWithLabels{
		RuntimeContextRuntimeID: runtimeCtxWithLabels,
	}

	appWithLabelsWithoutTemplate2 = &webhook.ApplicationWithLabels{
		Application: fixApplicationModelWithoutTemplate(Application2ID),
		Labels:      fixApplicationLabelsMap(),
	}

	listeningApplications = []*model.Application{{BaseEntity: &model.BaseEntity{ID: Application2ID}}, {ApplicationTemplateID: str.Ptr(ApplicationTemplateID), BaseEntity: &model.BaseEntity{ID: ApplicationID}}}

	applicationNotificationWithAppTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			Formation:   formationModelWithoutError,
			ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext:    nil,
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	applicationNotificationWithoutAppTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationWebhookGQLModel(Webhook2ID, Application2ID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:           model.AssignFormation,
			FormationID:         fixUUID(),
			ApplicationTemplate: nil,
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext:    nil,
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	runtimeNotificationWithRtmCtxAndAppTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixRuntimeWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext: &webhook.RuntimeContextWithLabels{
				RuntimeContext: fixRuntimeContextModel(),
				Labels:         fixRuntimeContextLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	appNotificationWithRtmCtxRtmIDAndTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
		Object: &webhook.FormationConfigurationChangeInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			Application: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			Runtime: &webhook.RuntimeWithLabels{
				Runtime: fixRuntimeModel(RuntimeID),
				Labels:  fixRuntimeLabelsMap(),
			},
			RuntimeContext: &webhook.RuntimeContextWithLabels{
				RuntimeContext: fixRuntimeContextModelWithRuntimeID(RuntimeID),
				Labels:         fixRuntimeContextLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	appToAppNotificationWithoutSourceTemplateWithTargetTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID),
		Object: &webhook.ApplicationTenantMappingInput{
			Operation:                 model.AssignFormation,
			FormationID:               fixUUID(),
			Formation:                 formationModelWithoutError,
			SourceApplicationTemplate: nil,
			SourceApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			TargetApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	appToAppNotificationWithSourceAndTargetTemplates = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID),
		Object: &webhook.ApplicationTenantMappingInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			SourceApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModelWithID(ApplicationTemplate2ID),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			TargetApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	appToAppNotificationWithSourceAndTargetTemplatesSwaped = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID),
		Object: &webhook.ApplicationTenantMappingInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			TargetApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModelWithID(ApplicationTemplate2ID),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			SourceApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	appToAppNotificationWithSourceTemplateWithoutTargetTemplate = &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID),
		Object: &webhook.ApplicationTenantMappingInput{
			Operation:   model.AssignFormation,
			FormationID: fixUUID(),
			SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixApplicationTemplateLabelsMap(),
			},
			SourceApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(ApplicationID),
				Labels:      fixApplicationLabelsMap(),
			},
			TargetApplicationTemplate: nil,
			TargetApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModelWithoutTemplate(Application2ID),
				Labels:      fixApplicationLabelsMap(),
			},
			Assignment:        emptyFormationAssignment,
			ReverseAssignment: emptyFormationAssignment,
		},
		CorrelationID: "",
	}

	// Formation notification variables
	emptyFormationNotificationRequests = make([]*webhookclient.FormationNotificationRequest, 0)

	formationNotificationSyncCreateRequest = &webhookclient.FormationNotificationRequest{
		Request: &webhookclient.Request{
			Webhook:       fixFormationLifecycleWebhookGQLModel(FormationLifecycleWebhookID, FormationTemplateID, graphql.WebhookModeSync),
			Object:        fixFormationLifecycleInput(model.CreateFormation, TntCustomerID, TntExternalID),
			CorrelationID: "",
		},
		Operation:     model.CreateFormation,
		Formation:     fixFormationModelWithoutError(),
		FormationType: testFormationTemplateName,
	}

	formationWithInitialState              = fixFormationModelWithState(model.InitialFormationState)
	formationNotificationSyncDeleteRequest = &webhookclient.FormationNotificationRequest{
		Request: &webhookclient.Request{
			Webhook:       fixFormationLifecycleWebhookGQLModel(FormationLifecycleWebhookID, FormationTemplateID, graphql.WebhookModeSync),
			Object:        fixFormationLifecycleInput(model.DeleteFormation, TntCustomerID, TntExternalID),
			CorrelationID: "",
		},
	}
	formationNotificationAsyncCreateRequest = &webhookclient.FormationNotificationRequest{
		Request: &webhookclient.Request{
			Webhook:       fixFormationLifecycleWebhookGQLModelAsync(FormationLifecycleWebhookID, FormationTemplateID),
			Object:        fixFormationLifecycleInput(model.CreateFormation, TntCustomerID, TntExternalID),
			CorrelationID: "",
		},
		Operation:     model.CreateFormation,
		Formation:     fixFormationModelWithoutError(),
		FormationType: testFormationTemplateName,
	}

	formationNotificationAsyncDeleteRequest = &webhookclient.FormationNotificationRequest{
		Request: &webhookclient.Request{
			Webhook:       fixFormationLifecycleWebhookGQLModelAsync(FormationLifecycleWebhookID, FormationTemplateID),
			Object:        fixFormationLifecycleInput(model.DeleteFormation, TntCustomerID, TntExternalID),
			CorrelationID: "",
		},
		Operation:     model.DeleteFormation,
		Formation:     fixFormationModelWithState(model.ReadyFormationState),
		FormationType: testFormationTemplateName,
	}

	formationNotificationSyncCreateRequests  = []*webhookclient.FormationNotificationRequest{formationNotificationSyncCreateRequest}
	formationNotificationSyncDeleteRequests  = []*webhookclient.FormationNotificationRequest{formationNotificationSyncDeleteRequest}
	formationNotificationAsyncDeleteRequests = []*webhookclient.FormationNotificationRequest{formationNotificationAsyncDeleteRequest}
	formationNotificationAsyncCreateRequests = []*webhookclient.FormationNotificationRequest{formationNotificationAsyncCreateRequest}

	formationNotificationWebhookSuccessResponse = fixFormationNotificationWebhookResponse(http.StatusOK, http.StatusOK, nil)
	formationNotificationWebhookErrorResponse   = fixFormationNotificationWebhookResponse(http.StatusOK, http.StatusOK, str.Ptr(testErr.Error()))

	formationLifecycleSyncWebhook  = fixFormationLifecycleSyncWebhookModel(FormationLifecycleWebhookID, FormationTemplateID, model.FormationTemplateWebhookReference)
	formationLifecycleSyncWebhooks = []*model.Webhook{formationLifecycleSyncWebhook}

	formationLifecycleAsyncWebhook  = fixFormationLifecycleAsyncCallbackWebhookModel(FormationLifecycleWebhookID, FormationTemplateID, model.FormationTemplateWebhookReference)
	formationLifecycleAsyncWebhooks = []*model.Webhook{formationLifecycleAsyncWebhook}
	emptyFormationLifecycleWebhooks []*model.Webhook

	// Formation constraints join point location variables
	preGenerateFormationAssignmentNotificationLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.GenerateFormationAssignmentNotificationOperation,
		ConstraintType: model.PreOperation,
	}

	postGenerateFormationAssignmentNotificationLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.GenerateFormationAssignmentNotificationOperation,
		ConstraintType: model.PostOperation,
	}

	preGenerateFormationNotificationLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.GenerateFormationNotificationOperation,
		ConstraintType: model.PreOperation,
	}

	postGenerateFormationNotificationLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.GenerateFormationNotificationOperation,
		ConstraintType: model.PostOperation,
	}

	preAssignLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.AssignFormationOperation,
		ConstraintType: model.PreOperation,
	}

	postAssignLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.AssignFormationOperation,
		ConstraintType: model.PostOperation,
	}

	preUnassignLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.UnassignFormationOperation,
		ConstraintType: model.PreOperation,
	}

	postUnassignLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.UnassignFormationOperation,
		ConstraintType: model.PostOperation,
	}

	preCreateLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.CreateFormationOperation,
		ConstraintType: model.PreOperation,
	}

	postCreateLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.CreateFormationOperation,
		ConstraintType: model.PostOperation,
	}

	preDeleteLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.DeleteFormationOperation,
		ConstraintType: model.PreOperation,
	}

	postDeleteLocation = formationconstraint.JoinPointLocation{
		OperationName:  model.DeleteFormationOperation,
		ConstraintType: model.PostOperation,
	}

	// Formation constraints join point details variables
	createFormationDetails = &formationconstraint.CRUDFormationOperationDetails{
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationName:       testFormationName,
		TenantID:            TntInternalID,
	}

	deleteFormationDetails = &formationconstraint.CRUDFormationOperationDetails{
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationName:       testFormationName,
		TenantID:            TntInternalID,
	}

	unassignAppDetails = &formationconstraint.UnassignFormationOperationDetails{
		ResourceType:        model.ApplicationResourceType,
		ResourceSubtype:     applicationType,
		ResourceID:          ApplicationID,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}

	assignAppInvalidTypeDetails = &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.ApplicationResourceType,
		ResourceSubtype:     "invalidApplicationType",
		ResourceID:          ApplicationID,
		FormationName:       secondTestFormationName,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}

	unassignRuntimeDetails = &formationconstraint.UnassignFormationOperationDetails{
		ResourceType:        model.RuntimeResourceType,
		ResourceSubtype:     runtimeType,
		ResourceID:          RuntimeID,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}

	assignRuntimeOtherTemplateDetails = &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.RuntimeResourceType,
		ResourceSubtype:     runtimeType,
		ResourceID:          RuntimeID,
		FormationName:       testFormationName,
		FormationType:       "some-other-template",
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}

	assignRuntimeContextOtherTemplateDetails = &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.RuntimeContextResourceType,
		ResourceSubtype:     runtimeType,
		ResourceID:          RuntimeContextID,
		FormationName:       testFormationName,
		FormationType:       "some-other-template",
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}

	unassignTenantDetails = &formationconstraint.UnassignFormationOperationDetails{
		ResourceType:        model.TenantResourceType,
		ResourceSubtype:     "account",
		ResourceID:          TargetTenant,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}

	notificationDetails = &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{}

	generateConfigurationChangeNotificationDetails = &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{
		Operation: model.AssignFormation,
		Formation: formationModelWithoutError,
		ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: fixApplicationTemplateModel(),
			Labels:              fixApplicationTemplateLabelsMap(),
		},
		Application: &webhook.ApplicationWithLabels{
			Application: fixApplicationModel(ApplicationID),
			Labels:      fixApplicationLabelsMap(),
		},
		Runtime: &webhook.RuntimeWithLabels{
			Runtime: fixRuntimeModel(RuntimeID),
			Labels:  fixRuntimeLabelsMap(),
		},
		Assignment:        emptyFormationAssignment,
		ReverseAssignment: emptyFormationAssignment,
	}

	generateAppToAppNotificationDetails = &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{
		Operation:                 model.AssignFormation,
		Formation:                 formationModelWithoutError,
		SourceApplicationTemplate: nil,
		SourceApplication: &webhook.ApplicationWithLabels{
			Application: fixApplicationModelWithoutTemplate(Application2ID),
			Labels:      fixApplicationLabelsMap(),
		},
		TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: fixApplicationTemplateModel(),
			Labels:              fixApplicationTemplateLabelsMap(),
		},
		TargetApplication: &webhook.ApplicationWithLabels{
			Application: fixApplicationModel(ApplicationID),
			Labels:      fixApplicationLabelsMap(),
		},
		Assignment:        emptyFormationAssignment,
		ReverseAssignment: emptyFormationAssignment,
	}

	formationNotificationDetails = &formationconstraint.GenerateFormationNotificationOperationDetails{
		Operation:             model.CreateFormation,
		FormationID:           FormationID,
		FormationName:         testFormationName,
		FormationType:         testFormationTemplateName,
		FormationTemplateID:   FormationTemplateID,
		TenantID:              TntInternalID,
		CustomerTenantContext: CustomerTenantContextAccount,
	}

	gaTenantObject = fixModelBusinessTenantMappingWithType(tnt.Account)
	rgTenantObject = fixModelBusinessTenantMappingWithType(tnt.ResourceGroup)

	customerTenantContext = &webhook.CustomerTenantContext{
		CustomerID: TntParentID,
		AccountID:  str.Ptr(gaTenantObject.ExternalTenant),
		Path:       nil,
	}

	rgCustomerTenantContext = &webhook.CustomerTenantContext{
		CustomerID: TntParentID,
		AccountID:  nil,
		Path:       str.Ptr(gaTenantObject.ExternalTenant),
	}
	firstFormationStatusParams  = dataloader.ParamFormationStatus{ID: FormationID, State: string(model.ReadyFormationState)}
	secondFormationStatusParams = dataloader.ParamFormationStatus{ID: FormationID + "2", State: string(model.InitialFormationState)}
	thirdFormationStatusParams  = dataloader.ParamFormationStatus{ID: FormationID + "3", State: string(model.ReadyFormationState)}
	fourthPageFormations        = dataloader.ParamFormationStatus{ID: FormationID + "4", State: string(model.ReadyFormationState)}
)

func unusedApplicationRepository() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

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

func unusedASAEngine() *automock.AsaEngine {
	return &automock.AsaEngine{}
}

func unusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func unusedTenantRepo() *automock.TenantRepository {
	return &automock.TenantRepository{}
}

func unusedRuntimeContextRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func unusedDataInputBuilder() *databuilderautomock.DataInputBuilder {
	return &databuilderautomock.DataInputBuilder{}
}

func expectEmptySliceRuntimeContextRepo() *automock.RuntimeContextRepository {
	repo := &automock.RuntimeContextRepository{}
	repo.On("ListByIDs", mock.Anything, TntInternalID, []string{}).Return(nil, nil).Once()
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

func unusedLabelDefRepository() *automock.LabelDefRepository {
	return &automock.LabelDefRepository{}
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

func unusedFormationAssignmentNotificationService() *automock.FormationAssignmentNotificationsService {
	return &automock.FormationAssignmentNotificationsService{}
}

func noActionNotificationsService() *automock.NotificationsService {
	notificationSvc := &automock.NotificationsService{}
	notificationSvc.On("GenerateFormationAssignmentNotifications", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	return notificationSvc
}

func unusedFormationTemplateRepo() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func unusedProcessFunc() *automock.ProcessFunc {
	return &automock.ProcessFunc{}
}

func unusedConstraintEngine() *automock.ConstraintEngine {
	return &automock.ConstraintEngine{}
}

func unusedNotificationsBuilder() *automock.NotificationBuilder {
	return &automock.NotificationBuilder{}
}

func unusedNotificationsGenerator() *automock.NotificationsGenerator {
	return &automock.NotificationsGenerator{}
}

func unusedTenantService() *automock.TenantService {
	return &automock.TenantService{}
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
	return []string{"id", "tenant_id", "formation_template_id", "name", "state", "error"}
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
	kind := model.RuntimeArtifactKindEnvironmentInstance
	return &model.FormationTemplate{
		ID:                     FormationTemplateID,
		Name:                   testFormationTemplateName,
		ApplicationTypes:       []string{"appType1", "appType2"},
		RuntimeTypes:           []string{"runtimeTypes"},
		RuntimeTypeDisplayName: str.Ptr("runtimeDisplayName"),
		RuntimeArtifactKind:    &kind,
	}
}

func fixFormationTemplateModelThatSupportsReset() *model.FormationTemplate {
	ftModel := fixFormationTemplateModel()
	ftModel.SupportsReset = true
	return ftModel
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

func fixApplicationModelWithTemplateID(applicationID, templateID string) *model.Application {
	app := fixApplicationModel(applicationID)
	app.ApplicationTemplateID = str.Ptr(templateID)
	return app
}

func fixModelBusinessTenantMappingWithType(t tnt.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             TntInternalID,
		Name:           "test-name",
		ExternalTenant: TntExternalID,
		Parent:         TntCustomerID,
		Type:           t,
		Provider:       testProvider,
		Status:         tnt.Active,
	}
}

func fixApplicationLabelsMap() map[string]string {
	return map[string]string{
		"app-label-key": "app-label-value",
	}
}

func fixApplicationTemplateLabelsMap() map[string]string {
	return map[string]string{
		"apptemplate-label-key": "apptemplate-label-value",
	}
}

func fixRuntimeLabelsMap() map[string]string {
	return map[string]string{
		"runtime-label-key": "runtime-label-value",
	}
}

func fixRuntimeContextLabelsMap() map[string]string {
	return map[string]string{
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

func fixFormationLifecycleWebhookModel(webhookID, objectID string, objectType model.WebhookReferenceObjectType, mode model.WebhookMode) *model.Webhook {
	return &model.Webhook{
		ID:         webhookID,
		ObjectID:   objectID,
		ObjectType: objectType,
		Type:       model.WebhookTypeFormationLifecycle,
		Mode:       &mode,
	}
}

func fixFormationLifecycleSyncWebhookModel(webhookID, objectID string, objectType model.WebhookReferenceObjectType) *model.Webhook {
	return fixFormationLifecycleWebhookModel(webhookID, objectID, objectType, model.WebhookModeSync)
}

func fixFormationLifecycleAsyncCallbackWebhookModel(webhookID, objectID string, objectType model.WebhookReferenceObjectType) *model.Webhook {
	return fixFormationLifecycleWebhookModel(webhookID, objectID, objectType, model.WebhookModeAsyncCallback)
}

func fixRuntimeWebhookGQLModel(webhookID, runtimeID string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:        webhookID,
		RuntimeID: str.Ptr(runtimeID),
		Type:      graphql.WebhookTypeConfigurationChanged,
	}
}

func fixRuntimeWebhookModel(webhookID, runtimeID string) *model.Webhook {
	return &model.Webhook{
		ID:       webhookID,
		ObjectID: runtimeID,
		Type:     model.WebhookTypeConfigurationChanged,
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
		ID:            webhookID,
		ApplicationID: str.Ptr(appID),
		Type:          graphql.WebhookTypeApplicationTenantMapping,
	}
}

func fixFormationLifecycleWebhookGQLModel(webhookID, formationTemplateID string, mode graphql.WebhookMode) graphql.Webhook {
	return graphql.Webhook{
		ID:                  webhookID,
		Type:                graphql.WebhookTypeFormationLifecycle,
		FormationTemplateID: &formationTemplateID,
		Mode:                &mode,
	}
}

func fixFormationLifecycleWebhookGQLModelAsync(webhookID, formationTemplateID string) graphql.Webhook {
	return fixFormationLifecycleWebhookGQLModel(webhookID, formationTemplateID, graphql.WebhookModeAsyncCallback)
}

func fixFormationLifecycleInput(formationOperation model.FormationOperation, customerTntID, accountTntExternalID string) *webhook.FormationLifecycleInput {
	return &webhook.FormationLifecycleInput{
		Operation:             formationOperation,
		Formation:             fixFormationModelWithoutError(),
		CustomerTenantContext: fixCustomerTenantContext(customerTntID, accountTntExternalID),
	}
}

func fixFormationNotificationWebhookResponse(actualStatusCode, successStatusCode int, err *string) *webhook.Response {
	return &webhook.Response{
		SuccessStatusCode: &successStatusCode,
		ActualStatusCode:  &actualStatusCode,
		Error:             err,
	}
}

func fixCustomerTenantContext(customerTntID, accountTntID string) *webhook.CustomerTenantContext {
	return &webhook.CustomerTenantContext{
		CustomerID: customerTntID,
		AccountID:  &accountTntID,
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

func fixApplicationTemplateModelWithID(id string) *model.ApplicationTemplate {
	template := fixApplicationTemplateModel()
	template.ID = id
	return template
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
		TenantID:            TntInternalID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
		State:               model.InitialFormationState,
		Error:               json.RawMessage(testFormationEmptyError),
	}
}

func fixFormationModelWithoutError() *model.Formation {
	return &model.Formation{
		ID:                  FormationID,
		TenantID:            TntInternalID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
}

func fixFormationModelWithState(state model.FormationState) *model.Formation {
	return &model.Formation{
		ID:                  FormationID,
		TenantID:            TntInternalID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
		State:               state,
	}
}

func fixFormationModelWithStateAndAssignmentError(t *testing.T, state model.FormationState, errMsg string, errCode formationassignment.AssignmentErrorCode) *model.Formation {
	formationError := formationassignment.AssignmentError{
		Message:   errMsg,
		ErrorCode: errCode,
	}

	marshaledErr, err := json.Marshal(formationError)
	require.NoError(t, err)

	return &model.Formation{
		ID:                  FormationID,
		TenantID:            TntInternalID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
		State:               state,
		Error:               marshaledErr,
	}
}

func fixFormationEntity() *formation.Entity {
	return &formation.Entity{
		ID:                  FormationID,
		TenantID:            TntInternalID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
		State:               string(model.InitialFormationState),
		Error:               repo.NewNullableStringFromJSONRawMessage(json.RawMessage("{}")),
	}
}

func fixGqlFormation() *graphql.Formation {
	return &graphql.Formation{
		ID:                  FormationID,
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		State:               string(model.ReadyFormationState),
	}
}

func fixGqlFormationAssignment(state string, configValue *string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:         FormationAssignmentID,
		Source:     FormationAssignmentSource,
		SourceType: FormationAssignmentSourceType,
		Target:     FormationAssignmentTarget,
		TargetType: FormationAssignmentTargetType,
		State:      state,
		Value:      configValue,
	}
}

func fixGqlFormationAssignmentWithSuffix(state string, configValue *string, suffix string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:         FormationAssignmentID + suffix,
		Source:     FormationAssignmentSource + suffix,
		SourceType: graphql.FormationAssignmentType(FormationAssignmentSourceType + suffix),
		Target:     FormationAssignmentTarget + suffix,
		TargetType: graphql.FormationAssignmentType(FormationAssignmentTargetType + suffix),
		State:      state,
		Value:      configValue,
	}
}

func fixFormationAssignmentModel(state string, configValue json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          FormationAssignmentID,
		FormationID: FormationAssignmentFormationID,
		TenantID:    FormationAssignmentTenantID,
		Source:      FormationAssignmentSource,
		SourceType:  FormationAssignmentSourceType,
		Target:      FormationAssignmentTarget,
		TargetType:  FormationAssignmentTargetType,
		State:       state,
		Value:       configValue,
	}
}

func fixFormationAssignmentModelWithParameters(id, formationID, source, target string, sourceType, targetType model.FormationAssignmentType, state model.FormationState) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          id,
		FormationID: formationID,
		Source:      source,
		SourceType:  sourceType,
		Target:      target,
		TargetType:  targetType,
		State:       string(state),
	}
}

func fixFormationAssignmentPairWithNoReverseAssignment(request *webhookclient.FormationAssignmentNotificationRequest, assignment *model.FormationAssignment) *formationassignment.AssignmentMappingPairWithOperation {
	res := &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			Assignment: &formationassignment.FormationAssignmentRequestMapping{
				Request:             request,
				FormationAssignment: assignment,
			},
			ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
				Request:             nil,
				FormationAssignment: nil,
			},
		},
	}

	switch assignment.State {
	case string(model.InitialAssignmentState), string(model.CreateErrorAssignmentState):
		res.Operation = model.AssignFormation
	case string(model.DeletingAssignmentState), string(model.DeleteErrorAssignmentState):
		res.Operation = model.UnassignFormation
	}

	return res
}

func fixFormationAssignmentModelWithSuffix(state string, configValue, errorValue json.RawMessage, suffix string) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          FormationAssignmentID + suffix,
		FormationID: FormationAssignmentFormationID + suffix,
		TenantID:    FormationAssignmentTenantID + suffix,
		Source:      FormationAssignmentSource + suffix,
		SourceType:  model.FormationAssignmentType(FormationAssignmentSourceType + suffix),
		Target:      FormationAssignmentTarget + suffix,
		TargetType:  model.FormationAssignmentType(FormationAssignmentTargetType + suffix),
		State:       state,
		Value:       configValue,
		Error:       errorValue,
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

func fixAssignAppDetails(formationName string) *formationconstraint.AssignFormationOperationDetails {
	return &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.ApplicationResourceType,
		ResourceSubtype:     applicationType,
		ResourceID:          ApplicationID,
		FormationName:       formationName,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}
}

func fixAssignRuntimeDetails(formationName string) *formationconstraint.AssignFormationOperationDetails {
	return &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.RuntimeResourceType,
		ResourceSubtype:     runtimeType,
		ResourceID:          RuntimeID,
		FormationName:       formationName,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}
}

func fixAssignRuntimeCtxDetails(formationName string) *formationconstraint.AssignFormationOperationDetails {
	return &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.RuntimeContextResourceType,
		ResourceSubtype:     runtimeType,
		ResourceID:          RuntimeContextID,
		FormationName:       formationName,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}
}

func fixAssignTenantDetails(formationName string) *formationconstraint.AssignFormationOperationDetails {
	return &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.TenantResourceType,
		ResourceSubtype:     "account",
		ResourceID:          TargetTenant,
		FormationName:       formationName,
		FormationType:       testFormationTemplateName,
		FormationTemplateID: FormationTemplateID,
		FormationID:         FormationID,
		TenantID:            TntInternalID,
	}
}

func fixDetailsForNotificationStatusReturned(formationType string, operation model.FormationOperation, location formationconstraint.JoinPointLocation, formation *model.Formation) *formationconstraint.NotificationStatusReturnedOperationDetails {
	return &formationconstraint.NotificationStatusReturnedOperationDetails{
		ResourceType:    model.FormationResourceType,
		ResourceSubtype: formationType,
		Location:        location,
		Operation:       operation,
		Formation:       formation,
	}
}
