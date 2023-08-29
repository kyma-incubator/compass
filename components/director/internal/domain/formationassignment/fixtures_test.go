package formationassignment_test

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
	TestFormationName       = "test-formation"
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
	TestErrorValueRawJSON         = json.RawMessage(`{"error":"error message"}`)
	TestEmptyErrorValueRawJSON    = json.RawMessage(`\"\"`)
	TestConfigValueStr            = "{\"configKey\":\"configValue\"}"
	TestErrorValueStr             = "{\"error\":\"error message\"}"
	fixColumns                    = []string{"id", "formation_id", "tenant_id", "source", "source_type", "target", "target_type", "state", "value", "error"}

	nilFormationAssignmentModel *model.FormationAssignment

	faModel                    = fixFormationAssignmentModel(TestConfigValueRawJSON)
	faModelWithConfigAndError  = fixFormationAssignmentModelWithConfigAndError(TestConfigValueRawJSON, TestErrorValueRawJSON)
	faEntityWithConfigAndError = fixFormationAssignmentEntityWithConfigurationAndError(TestConfigValueStr, TestErrorValueStr)

	appSubtype = "app-subtype"
	rtmSubtype = "rtm-subtype"
)

func fixFormationAssignmentGQLModel(configValue *string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:            TestID,
		Source:        TestSource,
		SourceType:    TestSourceType,
		Target:        TestTarget,
		TargetType:    TestTargetType,
		State:         TestStateInitial,
		Value:         configValue,
		Configuration: configValue,
		Error:         nil,
	}
}

func fixFormationAssignmentGQLModelWithError(errorValue *string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:            TestID,
		Source:        TestSource,
		SourceType:    TestSourceType,
		Target:        TestTarget,
		TargetType:    TestTargetType,
		State:         TestStateInitial,
		Value:         errorValue,
		Error:         errorValue,
		Configuration: nil,
	}
}

func fixFormationAssignmentGQLModelWithConfigAndError(configValue, errorValue *string) *graphql.FormationAssignment {
	return &graphql.FormationAssignment{
		ID:            TestID,
		Source:        TestSource,
		SourceType:    TestSourceType,
		Target:        TestTarget,
		TargetType:    TestTargetType,
		State:         TestStateInitial,
		Value:         errorValue,
		Error:         errorValue,
		Configuration: configValue,
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
		Error:       nil,
	}
}

func fixFormationAssignmentModelWithError(errorValue json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          TestID,
		FormationID: TestFormationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       nil,
		Error:       errorValue,
	}
}

func fixFormationAssignmentModelWithConfigAndError(configValue, errorValue json.RawMessage) *model.FormationAssignment {
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
		Error:       errorValue,
	}
}

func fixFormationAssignmentModelWithParameters(id, formationID, tenantID, sourceID, targetID string, sourceType, targetType model.FormationAssignmentType, state string, configValue, errorValue json.RawMessage) *model.FormationAssignment {
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
		Error:       errorValue,
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
		Error:       nil,
	}
}

func fixFormationAssignmentModelInputWithError(errorValue json.RawMessage) *model.FormationAssignmentInput {
	return &model.FormationAssignmentInput{
		FormationID: TestFormationID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       nil,
		Error:       errorValue,
	}
}

func fixFormationAssignmentModelInputWithConfigurationAndError(configValue, errorValue json.RawMessage) *model.FormationAssignmentInput {
	return &model.FormationAssignmentInput{
		FormationID: TestFormationID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       configValue,
		Error:       errorValue,
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
		Error:       sql.NullString{},
	}
}

func fixFormationAssignmentEntityWithError(errorValue string) *formationassignment.Entity {
	return &formationassignment.Entity{
		ID:          TestID,
		FormationID: TestFormationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       sql.NullString{},
		Error:       repo.NewValidNullableString(errorValue),
	}
}

func fixFormationAssignmentEntityWithConfigurationAndError(configValue, errorValue string) *formationassignment.Entity {
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
		Error:       repo.NewValidNullableString(errorValue),
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
		Formation:                 formation,
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

func fixAssignmentMappingPairWithID(id string) *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
				Request:             nil,
				FormationAssignment: &model.FormationAssignment{ID: id, Source: "source"},
			},
			ReverseAssignmentReqMapping: nil,
		},
		Operation: model.AssignFormation,
	}
}

func fixAssignmentMappingPairWithAssignmentAndRequest(assignment *model.FormationAssignment, req *webhookclient.FormationAssignmentNotificationRequest) *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
				Request:             req,
				FormationAssignment: assignment,
			},
			ReverseAssignmentReqMapping: nil,
		},
		Operation: model.AssignFormation,
	}
}

func fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(assignment, reverseAssignment *model.FormationAssignment, req, reverseReq *webhookclient.FormationAssignmentNotificationRequest) *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
				Request:             req,
				FormationAssignment: assignment,
			},
			ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
				Request:             reverseReq,
				FormationAssignment: reverseAssignment,
			},
		},
		Operation: model.AssignFormation,
	}
}

func fixExtendedFormationAssignmentNotificationReq(reqWebhook *webhookclient.FormationAssignmentNotificationRequest, fa *model.FormationAssignment) *webhookclient.FormationAssignmentNotificationRequestExt {
	return &webhookclient.FormationAssignmentNotificationRequestExt{
		FormationAssignmentNotificationRequest: reqWebhook,
		Operation:                              assignOperation,
		FormationAssignment:                    fa,
		ReverseFormationAssignment:             &model.FormationAssignment{},
		Formation:                              formation,
		TargetSubtype:                          appSubtype,
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
		Error:       assignment.Error,
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
		Error:       string(formationAssignment.Error),
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
		},
	}
}

func fixFormationAssignmentInputsWithObjectTypeAndID(objectType model.FormationAssignmentType, objectID, appID, rtmID, rtmCtxID string) []*model.FormationAssignmentInput {
	return []*model.FormationAssignmentInput{
		{
			FormationID: "ID",
			Source:      objectID,
			SourceType:  objectType,
			Target:      appID,
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      appID,
			SourceType:  model.FormationAssignmentTypeApplication,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      objectID,
			SourceType:  objectType,
			Target:      rtmID,
			TargetType:  model.FormationAssignmentTypeRuntime,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      rtmID,
			SourceType:  model.FormationAssignmentTypeRuntime,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      objectID,
			SourceType:  objectType,
			Target:      rtmCtxID,
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      rtmCtxID,
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		// Self formation assignments
		{
			FormationID: "ID",
			Source:      objectID,
			SourceType:  objectType,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
		},
	}
}

func fixFormationAssignmentInputsForSelf(appID, rtmID, rtmCtxID string) []*model.FormationAssignmentInput {
	return []*model.FormationAssignmentInput{
		{
			FormationID: "ID",
			Source:      appID,
			SourceType:  model.FormationAssignmentTypeApplication,
			Target:      appID,
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      rtmID,
			SourceType:  model.FormationAssignmentTypeRuntime,
			Target:      rtmID,
			TargetType:  model.FormationAssignmentTypeRuntime,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      rtmCtxID,
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      rtmCtxID,
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
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
			Error:       nil,
		},
	}
}

func fixFormationAssignmentInputsForRtmCtxWithAppAndRtmCtx(objectType model.FormationAssignmentType, objectID, appID, rtmCtxID string) []*model.FormationAssignmentInput {
	return []*model.FormationAssignmentInput{
		{
			FormationID: "ID",
			Source:      objectID,
			SourceType:  objectType,
			Target:      appID,
			TargetType:  model.FormationAssignmentTypeApplication,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      appID,
			SourceType:  model.FormationAssignmentTypeApplication,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      objectID,
			SourceType:  objectType,
			Target:      rtmCtxID,
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      rtmCtxID,
			SourceType:  model.FormationAssignmentTypeRuntimeContext,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.InitialAssignmentState),
			Value:       nil,
			Error:       nil,
		},
		{
			FormationID: "ID",
			Source:      objectID,
			SourceType:  objectType,
			Target:      objectID,
			TargetType:  objectType,
			State:       string(model.ReadyAssignmentState),
			Value:       nil,
			Error:       nil,
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

func fixNotificationStatusReturnedDetails(resourceType model.ResourceType, resourceSubtype string, fa, reverseFa *model.FormationAssignment, location formationconstraint.JoinPointLocation) *formationconstraint.NotificationStatusReturnedOperationDetails {
	return &formationconstraint.NotificationStatusReturnedOperationDetails{
		ResourceType:               resourceType,
		ResourceSubtype:            resourceSubtype,
		Location:                   location,
		Operation:                  assignOperation,
		FormationAssignment:        fa,
		ReverseFormationAssignment: reverseFa,
		Formation:                  formation,
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
	if formationAssignment == nil {
		return &webhook.FormationAssignment{Value: "\"\"", Error: "\"\""}
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
		Value:       str.StringifyJSONRawMessage(formationAssignment.Value),
		Error:       str.StringifyJSONRawMessage(formationAssignment.Error),
	}
}
