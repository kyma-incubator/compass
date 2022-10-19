package formation_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NotificationsService_GenerateNotifications(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("test error")

	inputFormation := model.Formation{
		Name: testFormationName,
	}
	expectedFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	testCases := []struct {
		Name                      string
		ApplicationRepoFN         func() *automock.ApplicationRepository
		ApplicationTemplateRepoFN func() *automock.ApplicationTemplateRepository
		RuntimeRepoFN             func() *automock.RuntimeRepository
		RuntimeContextRepoFn      func() *automock.RuntimeContextRepository
		LabelRepoFN               func() *automock.LabelRepository
		WebhookRepoFN             func() *automock.WebhookRepository
		WebhookConverterFN        func() *automock.WebhookConverter
		WebhookClientFN           func() *automock.WebhookClient
		ObjectID                  string
		ObjectType                graphql.FormationObjectType
		OperationType             model.FormationOperation
		InputFormation            model.Formation
		ExpectedRequests          []*webhookclient.NotificationRequest
		ExpectedErrMessage        string
	}{
		{
			Name: "success for runtime",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
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
						RuntimeContext: nil,
					},
					CorrelationID: "",
				},
				{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.AssignFormation,
						FormationID:         expectedFormation.ID,
						ApplicationTemplate: nil,
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				},
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime if webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application template labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application templates fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime if there are no applications to notify",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime if fetching applications fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching webhooks fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime labels fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime context",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:    graphql.FormationObjectTypeRuntimeContext,
			OperationType: model.AssignFormation,
			ObjectID:      RuntimeContextID,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
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
					},
					CorrelationID: "",
				},
				{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.AssignFormation,
						FormationID:         expectedFormation.ID,
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
					},
					CorrelationID: "",
				},
			},
			InputFormation:     expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime context if webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application template labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application templates fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching applications fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching webhook fails",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns nil for runtime context if webhook is not found",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.RuntimeWebhook, WebhookID))
				return repo
			},
			ObjectType:       graphql.FormationObjectTypeRuntimeContext,
			ObjectID:         RuntimeContextID,
			InputFormation:   inputFormation,
			ExpectedRequests: nil,
		},
		{
			Name: "error for runtime context if fetching runtime labels fails",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime fails",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context labels fails",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{}).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: templates list labels for IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: templates list by IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: application labels list for IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: application by scenarios and IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: self webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list labels for app templates of apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list app templates of apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list labels of apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success when there are no listening apps",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:     graphql.FormationObjectTypeApplication,
			ObjectID:       ApplicationID,
			InputFormation: inputFormation,
		},
		{
			Name: "error while generation app-to-app notifications: list listening apps' webhooks fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app's template labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app's template fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Once()
				repo.On("Get", ctx, ApplicationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app's labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Once()
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, testErr).Once()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Once()
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(nil, testErr).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Once()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for application with both runtime and app-to-app notifications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeID, RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{}).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookID, RuntimeID), fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID), fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
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
						RuntimeContext: nil,
					},
					CorrelationID: "",
				},
				{
					Webhook: *fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
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
					},
					CorrelationID: "",
				},
				{
					Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID),
					Object: &webhook.ApplicationTenantMappingInput{
						Operation:                 model.AssignFormation,
						FormationID:               expectedFormation.ID,
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
					},
					CorrelationID: "",
				},
				{
					Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID),
					Object: &webhook.ApplicationTenantMappingInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
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
					},
					CorrelationID: "",
				},
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application when webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime context labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime contexts in scenario fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtimes in scenario fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching webhooks fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFN != nil {
				runtimeRepo = testCase.RuntimeRepoFN()
			}
			runtimeContextRepo := unusedRuntimeContextRepo()
			if testCase.RuntimeContextRepoFn != nil {
				runtimeContextRepo = testCase.RuntimeContextRepoFn()
			}
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			webhookConverter := unusedWebhookConverter()
			if testCase.WebhookConverterFN != nil {
				webhookConverter = testCase.WebhookConverterFN()
			}
			appTemplateRepo := unusedAppTemplateRepository()
			if testCase.ApplicationTemplateRepoFN != nil {
				appTemplateRepo = testCase.ApplicationTemplateRepoFN()
			}
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFN != nil {
				labelRepo = testCase.LabelRepoFN()
			}

			notificationSvc := formation.NewNotificationService(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookConverter, nil)

			// WHEN
			actual, err := notificationSvc.GenerateNotifications(ctx, Tnt, testCase.ObjectID, &testCase.InputFormation, testCase.OperationType, testCase.ObjectType)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.ExpectedRequests, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, applicationRepo, webhookRepo, webhookConverter, appTemplateRepo, labelRepo)
		})
	}
}

func Test_NotificationsService_SendNotification(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("test error")

	testCases := []struct {
		Name               string
		WebhookClientFN    func() *automock.WebhookClient
		InputRequest       *webhookclient.NotificationRequest
		ExpectedErrMessage string
	}{
		{
			Name: "success when webhook client call doesn't return error",
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.NotificationRequest{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: FormationID,
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
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil)
				return client
			},
			InputRequest: &webhookclient.NotificationRequest{

				Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
				Object: &webhook.FormationConfigurationChangeInput{
					Operation:   model.AssignFormation,
					FormationID: FormationID,
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
					RuntimeContext: nil,
				},
				CorrelationID: "",
			},
		},
		{
			Name: "success when webhook client call returns error from error field in response",
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.NotificationRequest{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: FormationID,
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
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(&webhook.Response{Error: str.Ptr(testErr.Error())}, testErr)
				return client
			},
			InputRequest: &webhookclient.NotificationRequest{

				Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
				Object: &webhook.FormationConfigurationChangeInput{
					Operation:   model.AssignFormation,
					FormationID: FormationID,
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
					RuntimeContext: nil,
				},
				CorrelationID: "",
			},
		},
		{
			Name: "fail when webhook client call fails",
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.NotificationRequest{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: FormationID,
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
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, testErr)
				return client
			},
			InputRequest: &webhookclient.NotificationRequest{
				Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
				Object: &webhook.FormationConfigurationChangeInput{
					Operation:   model.AssignFormation,
					FormationID: FormationID,
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
					RuntimeContext: nil,
				},
				CorrelationID: "",
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "does nothing when no arguments are supplied",
			InputRequest: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			webhookClient := unusedWebhookClient()
			if testCase.WebhookClientFN != nil {
				webhookClient = testCase.WebhookClientFN()
			}

			notificationSvc := formation.NewNotificationService(nil, nil, nil, nil, nil, nil, nil, webhookClient)

			// WHEN
			_, err := notificationSvc.SendNotification(ctx, testCase.InputRequest)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}
