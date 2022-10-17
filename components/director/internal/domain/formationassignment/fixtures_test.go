package formationassignment_test

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

const (
	TestID          = "c861c3db-1265-4143-a05c-1ced1291d816"
	TestFormationID = "a7c0bd01-2441-4ca1-9b5e-a54e74fd7773"
	TestTenantID    = "b4d1bd32-dd07-4141-9655-42bc33a4ae37"
	TestSource      = "05e10560-2259-4adf-bb3e-6aee0518f573"
	TestSourceType  = "application"
	TestTarget      = "1c22035a-72e4-4a78-9025-bbcb1f87760b"
	TestTargetType  = "runtimeContext"
	TestState       = "INITIAL"
)

var (
	TestConfigValueRawJSON        = json.RawMessage(`{"configKey":"configValue"}`)
	TestInvalidConfigValueRawJSON = json.RawMessage(`{invalid}`)
	TestConfigValueStr            = "{\"configKey\":\"configValue\"}"
	fixColumns                    = []string{"id", "formation_id", "tenant_id", "source", "source_type", "target", "target_type", "state", "value"}

	nilFormationAssignmentModel *model.FormationAssignment

	faModel  = fixFormationAssignmentModel(TestConfigValueRawJSON)
	faEntity = fixFormationAssignmentEntity(TestConfigValueStr)
)

func fixFormationAssignmentGQLModel(configValue *string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:         TestID,
		Source:     TestSource,
		SourceType: TestSourceType,
		Target:     TestTarget,
		TargetType: TestTargetType,
		State:      TestState,
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
		State:       TestState,
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
		State:       TestState,
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
		State:       TestState,
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
		State:       TestState,
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
		State:       TestState,
		Value:       repo.NewValidNullableString(TestConfigValueStr),
	}
}

func fixUUIDService() *automock.UIDService {
	uidSvc := &automock.UIDService{}
	uidSvc.On("Generate").Return(TestID)
	return uidSvc
}

func fixFormationAssignment() *model.FormationAssignment {
	return &model.FormationAssignment{Source: "source"}
}

func fixFormationAssignmentWithID(id string) *model.FormationAssignment {
	return &model.FormationAssignment{ID: id, Source: "source"}
}

func fixAssignmentMappingPairWithID(id string) *formationassignment.AssignmentMappingPair {
	return &formationassignment.AssignmentMappingPair{
		Assignment: &formationassignment.FormationAssignmentRequestMapping{
			Request:             nil,
			FormationAssignment: &model.FormationAssignment{ID: id, Source: "source"},
		},
		ReverseAssignment: nil,
	}
}

func fixAssignmentMappingPairWithIDAndRequest(id string, req *webhookclient.NotificationRequest) *formationassignment.AssignmentMappingPair {
	return &formationassignment.AssignmentMappingPair{
		Assignment: &formationassignment.FormationAssignmentRequestMapping{
			Request:             req,
			FormationAssignment: &model.FormationAssignment{ID: id, Source: "source"},
		},
		ReverseAssignment: nil,
	}
}

func fixFormationAssignmentsWithObjectTypeAndID(objectType graphql.FormationObjectType, objectID, appID, rtmID, rtmCtxID string) []*model.FormationAssignment {
	return []*model.FormationAssignment{
		&model.FormationAssignment{
			ID:          "ID1",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      appID,
			TargetType:  string(graphql.FormationObjectTypeApplication),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID2",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      appID,
			SourceType:  string(graphql.FormationObjectTypeApplication),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID3",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      rtmID,
			TargetType:  string(graphql.FormationObjectTypeRuntime),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID4",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmID,
			SourceType:  string(graphql.FormationObjectTypeRuntime),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID5",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      rtmCtxID,
			TargetType:  string(graphql.FormationObjectTypeRuntimeContext),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID6",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmCtxID,
			SourceType:  string(graphql.FormationObjectTypeRuntimeContext),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
	}
}

func fixFormationAssignmentsForRtmCtxWithAppAndRtmCtx(objectType graphql.FormationObjectType, objectID, appID, rtmCtxID string) []*model.FormationAssignment {
	return []*model.FormationAssignment{
		&model.FormationAssignment{
			ID:          "ID1",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      appID,
			TargetType:  string(graphql.FormationObjectTypeApplication),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID2",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      appID,
			SourceType:  string(graphql.FormationObjectTypeApplication),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID3",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      rtmCtxID,
			TargetType:  string(graphql.FormationObjectTypeRuntimeContext),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			ID:          "ID4",
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      rtmCtxID,
			SourceType:  string(graphql.FormationObjectTypeRuntimeContext),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
	}
}

func fixNotificationRequestAndReverseRequest(objectID, object2ID string, participants []string, assignment, assignmentReverse *model.FormationAssignment, webhookType, reverseWebhookType string, hasReverseWebhook bool) ([]*webhookclient.NotificationRequest, *automock.TemplateInput, *automock.TemplateInput) {
	var request *webhookclient.NotificationRequest
	var requestReverse *webhookclient.NotificationRequest

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

	request = &webhookclient.NotificationRequest{Webhook: webhook, Object: templateInput}

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

		requestReverse = &webhookclient.NotificationRequest{Webhook: webhookReverse, Object: templateInputReverse}
	} else {
		requestReverse = nil
	}

	return []*webhookclient.NotificationRequest{request, requestReverse}, templateInput, templateInputReverse
}
