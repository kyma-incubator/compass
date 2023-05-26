package formationassignment_test

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"

	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	databuilderautomock "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

const (
	TestID                  = "c861c3db-1265-4143-a05c-1ced1291d816"
	TestFormationID         = "a7c0bd01-2441-4ca1-9b5e-a54e74fd7773"
	TestFormationTemplateID = "jjc0bd01-2441-4ca1-9b5e-a54e74fd7773"
	TestTenantID            = "b4d1bd32-dd07-4141-9655-42bc33a4ae37"
	TestSource              = "05e10560-2259-4adf-bb3e-6aee0518f573"
	TestSourceType          = "application"
	TestTarget              = "1c22035a-72e4-4a78-9025-bbcb1f87760b"
	TestTargetType          = "runtimeContext"
	TestStateInitial        = "INITIAL"
	TestReadyState          = "READY"
	TestWebhookID           = "eca98d44-aac0-4e44-898b-c394beab2e94"
	TestReverseWebhookID    = "aecec253-b4d8-416a-be5c-a27677ee5157"
	TntParentID             = "2d11035a-72e4-4a78-9025-bbcb1f87760b"
)

var (
	TestConfigValueRawJSON        = json.RawMessage(`{"configKey":"configValue"}`)
	TestInvalidConfigValueRawJSON = json.RawMessage(`{invalid}`)
	TestConfigValueStr            = "{\"configKey\":\"configValue\"}"
	fixColumns                    = []string{"id", "formation_id", "tenant_id", "source", "source_type", "target", "target_type", "state", "value"}

	nilFormationAssignmentModel *model.FormationAssignment

	faModel  = fixFormationAssignmentModel(TestConfigValueRawJSON)
	faEntity = fixFormationAssignmentEntity(TestConfigValueStr)

	appSubtype = "subtype"
)

func fixFormationAssignmentGQLModel(configValue *string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:         TestID,
		Source:     TestSource,
		SourceType: TestSourceType,
		Target:     TestTarget,
		TargetType: TestTargetType,
		State:      TestStateInitial,
		Value:      configValue,
	}
}

func fixFormationAssignmentModel(configValue json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          TestID,
		FormationID: TestFormationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       configValue,
	}
}

func fixFormationAssignmentModelWithParameters(id, formationID, tenantID, sourceID, targetID string, sourceType, targetType model.FormationAssignmentType, state string, configValue json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          id,
		FormationID: formationID,
		TenantID:    tenantID,
		Source:      sourceID,
		SourceType:  sourceType,
		Target:      targetID,
		TargetType:  targetType,
		State:       state,
		Value:       configValue,
	}
}

func fixFormationAssignmentModelWithFormationID(formationID string) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          TestID,
		FormationID: formationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       TestConfigValueRawJSON,
	}
}

func fixFormationAssignmentModelWithIDAndTenantID(fa *model.FormationAssignment) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          TestID,
		FormationID: fa.FormationID,
		TenantID:    TestTenantID,
		Source:      fa.Source,
		SourceType:  fa.SourceType,
		Target:      fa.Target,
		TargetType:  fa.TargetType,
		State:       fa.State,
		Value:       fa.Value,
	}
}

func fixFormationAssignmentModelInput(configValue json.RawMessage) *model.FormationAssignmentInput {
	return &model.FormationAssignmentInput{
		FormationID: TestFormationID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       configValue,
	}
}

func fixFormationAssignmentEntity(configValue string) *formationassignment.Entity {
	return &formationassignment.Entity{
		ID:          TestID,
		FormationID: TestFormationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       repo.NewValidNullableString(configValue),
	}
}

func fixFormationAssignmentEntityWithFormationID(formationID string) *formationassignment.Entity {
	return &formationassignment.Entity{
		ID:          TestID,
		FormationID: formationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       repo.NewValidNullableString(TestConfigValueStr),
	}
}

func fixAppTenantMappingWebhookInput(formationID string, sourceApp, targetApp *webhook.ApplicationWithLabels, sourceAppTemplate, targetAppTemplate *webhook.ApplicationTemplateWithLabels, customerTenantContext *webhook.CustomerTenantContext, assignment, reverseAssignment *webhook.FormationAssignment) *webhook.ApplicationTenantMappingInput {
	return &webhook.ApplicationTenantMappingInput{
		Operation:                 model.AssignFormation,
		FormationID:               formationID,
		SourceApplicationTemplate: sourceAppTemplate,
		SourceApplication:         sourceApp,
		TargetApplicationTemplate: targetAppTemplate,
		TargetApplication:         targetApp,
		CustomerTenantContext:     customerTenantContext,
		Assignment:                assignment,
		ReverseAssignment:         reverseAssignment,
	}
}

func fixFormationConfigurationChangeInput(formationID string, appTemplate *webhook.ApplicationTemplateWithLabels, app *webhook.ApplicationWithLabels, runtime *webhook.RuntimeWithLabels, runtimeCtx *webhook.RuntimeContextWithLabels, customerTenantContext *webhook.CustomerTenantContext, assignment, reverseAssignment *webhook.FormationAssignment) *webhook.FormationConfigurationChangeInput {
	return &webhook.FormationConfigurationChangeInput{
		Operation:             model.AssignFormation,
		FormationID:           formationID,
		ApplicationTemplate:   appTemplate,
		Application:           app,
		Runtime:               runtime,
		RuntimeContext:        runtimeCtx,
		CustomerTenantContext: customerTenantContext,
		Assignment:            assignment,
		ReverseAssignment:     reverseAssignment,
	}
}

func fixModelBusinessTenantMappingWithType(t tnt.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             TestTenantID,
		Name:           "test-name",
		ExternalTenant: TestTenantID,
		Parent:         TntParentID,
		Type:           t,
		Provider:       "Compass",
		Status:         tnt.Active,
	}
}

func fixFormationAssignmentOnlyWithSourceAndTarget() *model.FormationAssignment {
	return &model.FormationAssignment{Source: "source", Target: "target"}
}

func fixAssignmentMappingPairWithID(id string) *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			Assignment: &formationassignment.FormationAssignmentRequestMapping{
				Request:             nil,
				FormationAssignment: &model.FormationAssignment{ID: id, Source: "source"},
			},
			ReverseAssignment: nil,
		},
		Operation: model.AssignFormation,
	}
}

func fixAssignmentMappingPairWithIDAndRequest(id string, req *webhookclient.FormationAssignmentNotificationRequest) *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			Assignment: &formationassignment.FormationAssignmentRequestMapping{
				Request:             req,
				FormationAssignment: &model.FormationAssignment{ID: id, Source: "source"},
			},
			ReverseAssignment: nil,
		},
		Operation: model.AssignFormation,
	}
}

func fixAssignmentMappingPairWithAssignmentAndRequest(assignment *model.FormationAssignment, req *webhookclient.FormationAssignmentNotificationRequest) *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			Assignment: &formationassignment.FormationAssignmentRequestMapping{
				Request:             req,
				FormationAssignment: assignment,
			},
			ReverseAssignment: nil,
		},
		Operation: model.AssignFormation,
	}
}

func fixExtendedFormationAssignmentNotificationReq(reqWebhook *webhookclient.FormationAssignmentNotificationRequest, fa *model.FormationAssignment) *formationassignment.FormationAssignmentRequestExt {
	return &formationassignment.FormationAssignmentRequestExt{
		FormationAssignmentNotificationRequest: reqWebhook,
		Operation:                              assignOperation,
		FormationAssignment:                    fa,
		Formation:                              formation,
		TargetSubtype:                          appSubtype,
	}
}

func fixFormationAssignmentWithConfigAndState(assignment *model.FormationAssignment, state model.FormationAssignmentState, value json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          assignment.ID,
		FormationID: assignment.FormationID,
		TenantID:    assignment.TenantID,
		Source:      assignment.Source,
		SourceType:  assignment.SourceType,
		Target:      assignment.Target,
		TargetType:  assignment.TargetType,
		State:       string(state),
		Value:       value,
	}
}

func fixFormationAssignmentWithConfigAndStateInput(assignment *model.FormationAssignmentInput, state model.FormationAssignmentState, value json.RawMessage) *model.FormationAssignmentInput {
	return &model.FormationAssignmentInput{
		FormationID: assignment.FormationID,
		Source:      assignment.Source,
		SourceType:  assignment.SourceType,
		Target:      assignment.Target,
		TargetType:  assignment.TargetType,
		State:       string(state),
		Value:       value,
	}
}

func fixReverseFormationAssignment(assignment *model.FormationAssignment) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          assignment.ID,
		FormationID: assignment.FormationID,
		TenantID:    assignment.TenantID,
		Source:      assignment.Target,
		SourceType:  assignment.TargetType,
		Target:      assignment.Source,
		TargetType:  assignment.SourceType,
		State:       assignment.State,
		Value:       assignment.Value,
	}
}

func fixConvertFAFromModel(formationAssignment *model.FormationAssignment) *webhook.FormationAssignment {
	return &webhook.FormationAssignment{
		ID:          formationAssignment.ID,
		FormationID: formationAssignment.FormationID,
		TenantID:    formationAssignment.TenantID,
		Source:      formationAssignment.Source,
		SourceType:  formationAssignment.SourceType,
		Target:      formationAssignment.Target,
		TargetType:  formationAssignment.TargetType,
		State:       formationAssignment.State,
		Value:       string(formationAssignment.Value),
	}
}

func fixFormationAssignmentsWithObjectTypeAndID(objectType model.FormationAssignmentType, objectID, appID, rtmID, rtmCtxID string) []*model.FormationAssignment {
	return []*model.FormationAssignment{
		{
			ID:          "ID1",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  objectType,
			Target:      appID,
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID2",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      appID,
			SourceType:  model.FormationAssignmentTypeApplication,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID3",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  objectType,
			Target:      rtmID,
			TargetType:  model.FormationAssignmentTypeRuntime,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID4",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmID,
			SourceType:  model.FormationAssignmentTypeRuntime,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID5",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  objectType,
			Target:      rtmCtxID,
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID6",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmCtxID,
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		// Self formation assignments
		{
			ID:          "ID7",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  objectType,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
		},
	}
}

func fixFormationAssignmentsForSelf(appID, rtmID, rtmCtxID string) []*model.FormationAssignment {
	return []*model.FormationAssignment{
		{
			ID:          "ID8",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      appID,
			SourceType:  model.FormationAssignmentTypeApplication,
			Target:      appID,
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID9",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmID,
			SourceType:  model.FormationAssignmentTypeRuntime,
			Target:      rtmID,
			TargetType:  model.FormationAssignmentTypeRuntime,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID10",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmCtxID,
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      rtmCtxID,
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
		},
	}
}

func fixFormationAssignmentsForRtmCtxWithAppAndRtmCtx(objectType model.FormationAssignmentType, objectID, appID, rtmCtxID string) []*model.FormationAssignment {
	return []*model.FormationAssignment{
		{
			ID:          "ID1",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  objectType,
			Target:      appID,
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID2",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      appID,
			SourceType:  model.FormationAssignmentTypeApplication,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID3",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  objectType,
			Target:      rtmCtxID,
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID4",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmCtxID,
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		{
			ID:          "ID5",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  objectType,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
		},
	}
}

func fixNotificationRequestAndReverseRequest(objectID, object2ID string, participants []string, assignment, assignmentReverse *model.FormationAssignment, webhookType, reverseWebhookType string, hasReverseWebhook bool) ([]*webhookclient.FormationAssignmentNotificationRequest, *automock.TemplateInput, *automock.TemplateInput) {
	var request *webhookclient.FormationAssignmentNotificationRequest
	var requestReverse *webhookclient.FormationAssignmentNotificationRequest

	templateInput := &automock.TemplateInput{}
	templateInputReverse := &automock.TemplateInput{}

	webhook := graphql.Webhook{}
	webhookReverse := graphql.Webhook{}
	switch webhookType {
	case "application":
		webhook.ApplicationID = &objectID
	case "runtime":
		webhook.RuntimeID = &objectID
	}

	templateInput.Mock.On("GetParticipantsIDs").Return(participants).Times(1)
	templateInput.Mock.On("SetAssignment", assignment).Times(2)
	templateInput.Mock.On("SetReverseAssignment", assignmentReverse).Times(2)

	request = &webhookclient.FormationAssignmentNotificationRequest{Webhook: webhook, Object: templateInput}

	if hasReverseWebhook {
		switch reverseWebhookType {
		case "application":
			webhookReverse.ApplicationID = &object2ID
		case "runtime":
			webhookReverse.RuntimeID = &object2ID
		}

		templateInputReverse.Mock.On("GetParticipantsIDs").Return(participants).Times(1)
		templateInputReverse.Mock.On("SetAssignment", assignmentReverse).Times(2)
		templateInputReverse.Mock.On("SetReverseAssignment", assignment).Times(2)

		requestReverse = &webhookclient.FormationAssignmentNotificationRequest{Webhook: webhookReverse, Object: templateInputReverse}
	} else {
		requestReverse = nil
	}

	return []*webhookclient.FormationAssignmentNotificationRequest{request, requestReverse}, templateInput, templateInputReverse
}

func fixNotificationStatusReturnedDetails(fa, reverseFa *model.FormationAssignment, location formationconstraint.JoinPointLocation) *formationconstraint.NotificationStatusReturnedOperationDetails {
	return &formationconstraint.NotificationStatusReturnedOperationDetails{
		ResourceType:               model.FormationResourceType,
		ResourceSubtype:            formationTemplate.Name,
		Location:                   location,
		Operation:                  assignOperation,
		FormationAssignment:        fa,
		ReverseFormationAssignment: reverseFa,
		Formation:                  formation,
		FormationTemplate:          formationTemplate,
	}
}

func fixUUIDService() *automock.UIDService {
	uidSvc := &automock.UIDService{}
	uidSvc.On("Generate").Return(TestID)
	return uidSvc
}

func unusedFormationAssignmentRepository() *automock.FormationAssignmentRepository {
	return &automock.FormationAssignmentRepository{}
}

func unusedUIDService() *automock.UIDService {
	return &automock.UIDService{}
}

func unusedNotificationService() *automock.NotificationService {
	return &automock.NotificationService{}
}

func unusedRuntimeRepository() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func unusedRuntimeContextRepository() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func unusedWebhookDataInputBuilder() *databuilderautomock.DataInputBuilder {
	return &databuilderautomock.DataInputBuilder{}
}

func unusedWebhookRepo() *automock.WebhookRepository {
	return &automock.WebhookRepository{}
}

func unusedWebhookConverter() *automock.WebhookConverter {
	return &automock.WebhookConverter{}
}

func unusedFormationRepo() *automock.FormationRepository {
	return &automock.FormationRepository{}
}

func unusedTenantRepo() *automock.TenantRepository {
	return &automock.TenantRepository{}
}

func unusedNotificationBuilder() *automock.NotificationBuilder {
	return &automock.NotificationBuilder{}
}

func convertFormationAssignmentFromModel(formationAssignment *model.FormationAssignment) *webhook.FormationAssignment {
	config := string(formationAssignment.Value)
	if config == "" {
		config = "\"\""
	}
	return &webhook.FormationAssignment{
		ID:          formationAssignment.ID,
		FormationID: formationAssignment.FormationID,
		TenantID:    formationAssignment.TenantID,
		Source:      formationAssignment.Source,
		SourceType:  formationAssignment.SourceType,
		Target:      formationAssignment.Target,
		TargetType:  formationAssignment.TargetType,
		State:       formationAssignment.State,
		Value:       config,
	}
}
