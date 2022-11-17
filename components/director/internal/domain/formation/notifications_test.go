package formation_test

import (
	"context"
	"testing"

	databuilderautomock "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder/automock"

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

	runtimeWithLabels := fixRuntimeWithLabels(RuntimeID)

	runtimeCtxWithLabels := &webhook.RuntimeContextWithLabels{
		RuntimeContext: fixRuntimeContextModel(),
		Labels:         fixRuntimeContextLabelsMap(),
	}

	appTemplateWithLabels := &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: fixApplicationTemplateModel(),
		Labels:              fixApplicationTemplateLabelsMap(),
	}

	appWithLabels := &webhook.ApplicationWithLabels{
		Application: fixApplicationModel(ApplicationID),
		Labels:      fixApplicationLabelsMap(),
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
		DataInputBuilder          func() *databuilderautomock.DataInputBuilder
		ObjectID                  string
		ObjectType                graphql.FormationObjectType
		OperationType             model.FormationOperation
		InputFormation            model.Formation
		ExpectedRequests          []*webhookclient.NotificationRequest
		ExpectedErrMessage        string
	}{
		// start testing 'generateNotificationsAboutApplicationsForTheRuntimeThatIsAssigned' and 'generateNotificationsForApplicationsAboutTheRuntimeThatIsAssigned' funcs
		{
			Name: "success when generating notifications for runtime about all applications in that formation and" +
				"success when generating notifications for all listening applications about the assigned runtime in that formation",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)

				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)

				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
					Webhook: *fixApplicationWebhookGQLModel(Webhook2ID, Application2ID),
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
			Name: "error when generating notifications for application if webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", mock.MatchedBy(func(wh *model.Webhook) bool { return checkIfIDInSet(wh, []string{WebhookID, Webhook2ID}) })).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application if fetching application template labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application if fetching application templates fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application if fetching application labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success when generating notifications for runtime if there are no applications in the formation to notify",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
			Name: "success when generating notifications for runtime if there are no applications with CONFIGURATION_CHANGED webhook to notify",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
			Name: "error when generating notifications for application if fetching applications fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application if fetching webhooks fails",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return(nil, testErr)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application if preparing runtime with labels fails",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime if webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(nil, testErr)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime if fetching application template labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime if fetching application templates fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime if fetching application labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns nil when generating notifications for runtime if webhook is not found",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.RuntimeWebhook, WebhookID))
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
					Webhook: *fixApplicationWebhookGQLModel(Webhook2ID, Application2ID),
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
		},
		{
			Name: "success when generating notifications for runtime if there are no applications to notify",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntime,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
					Webhook: *fixApplicationWebhookGQLModel(Webhook2ID, Application2ID),
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
			Name: "error when generating notifications for runtime if fetching applications fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime if fetching webhooks fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime if preparing runtime with labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(runtimeWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		// start testing 'generateNotificationsForApplicationsAboutTheRuntimeContextThatIsAssigned' and 'generateNotificationsAboutApplicationsForTheRuntimeContextThatIsAssigned' funcs
		{
			Name: "success when generating notifications for runtime contexts about all applications in that formation",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)

				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)

				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:    graphql.FormationObjectTypeRuntimeContext,
			OperationType: model.AssignFormation,
			ObjectID:      RuntimeContextID,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
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
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
					Webhook: *fixApplicationWebhookGQLModel(Webhook2ID, Application2ID),
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
			Name: "error when generating notifications for application  when runtime context is assigned if webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when runtime context is assigned if fetching application template with labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when runtime context is assigned if fetching application templates fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when runtime context is assigned if fetching application labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success when generating notifications for runtime context if there are no applications in the formation to notify",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntimeContext,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeContextID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
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
			ExpectedErrMessage: "",
		},
		{
			Name: "success when generating notifications for runtime context if there are no applications with CONFIGURATION_CHANGED webhook to notify",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntimeContext,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeContextID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
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
			ExpectedErrMessage: "",
		},
		{
			Name: "error when generating notifications for application when runtime context is assigned if fetching applications fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when runtime context is assigned if fetching webhook fails",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return(nil, testErr)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when runtime context is assigned if preparing runtime context with labels fails",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when runtime context is assigned if preparing runtime with labels fails",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime context if webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(nil, testErr)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime context if fetching application template labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime context if fetching application templates fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime context if fetching application labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime context if fetching applications fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, RuntimeContextRuntimeID, model.RuntimeWebhookReference), nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime context if fetching webhook fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "returns nil when generating notifications for runtime context if webhook is not found",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.RuntimeWebhook, WebhookID))
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Twice()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Twice()
				return dataInputBuilder
			},
			ObjectType:     graphql.FormationObjectTypeRuntimeContext,
			OperationType:  model.AssignFormation,
			ObjectID:       RuntimeContextID,
			InputFormation: expectedFormation,
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
					Webhook: *fixApplicationWebhookGQLModel(Webhook2ID, Application2ID),
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
		},
		{
			Name: "error when generating notifications for runtime context if fetching runtime context with labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for runtime context if fetching runtime with labels fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{expectedFormation.Name}, []string{ApplicationID, Application2ID}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(Webhook2ID, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{ApplicationID, Application2ID}) })).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.ApplicationWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), fixConfigurationChangedWebhookModel(Webhook2ID, Application2ID, model.ApplicationWebhookReference)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeWithLabels(RuntimeContextRuntimeID), nil).Once()
				dataInputBuilder.On("PrepareRuntimeContextWithLabels", ctx, Tnt, RuntimeContextID).Return(runtimeCtxWithLabels, nil).Once()
				dataInputBuilder.On("PrepareRuntimeWithLabels", ctx, Tnt, RuntimeContextRuntimeID).Return(nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		// start testing 'generateRuntimeNotificationsForApplicationAssignment' and 'generateApplicationNotificationsForApplicationAssignment' funcs
		{
			Name: "success when generating notifications for application with both runtime <-> app and app <-> app notifications",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
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
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Twice()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
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
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModelWithRuntimeID(RuntimeID),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				},
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			OperationType:      model.AssignFormation,
			ObjectID:           ApplicationID,
			InputFormation:     expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error while generating app-to-app notifications: webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)).Return(nil, testErr)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: templates list labels for IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: templates list by IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: application labels list for IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: application by scenarios and IDs fail",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: self webhook conversion fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", mock.MatchedBy(func(wh *model.Webhook) bool {
					return checkIfIDInSet(wh, []string{AppTenantMappingWebhookIDForApp1, AppTenantMappingWebhookIDForApp2})
				})).Return(nil, testErr)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: list labels for app templates of apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: list app templates of apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: list labels of apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: list apps already in formation fails",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success when there are no listening apps",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			OperationType:      model.AssignFormation,
			InputFormation:     expectedFormation,
			ExpectedErrMessage: "",
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModelWithRuntimeID(RuntimeID),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				},
				{
					Webhook: *fixApplicationWebhookGQLModel(WebhookID, ApplicationID),
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
			},
		},
		{
			Name: "error while generating app-to-app notifications: list listening apps' webhooks fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, testErr)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generating app-to-app notifications: while preparing app template with labels fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(2)
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when webhook conversion fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
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
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(nil, testErr)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime context labels fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime contexts in scenario fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtimes in scenario fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes labels fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching webhooks fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, testErr)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Twice()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching app and/or application template with labels fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(fixApplicationWebhookGQLModel(WebhookID, ApplicationID), nil).Twice()
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(nil, nil, testErr).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when application is assigned if webhook conversion fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(map[string]map[string]interface{}{
					RuntimeContextID:  fixRuntimeContextLabelsMap(),
					RuntimeContext2ID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference)).Return(nil, testErr)
				return conv
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when application is assigned if fetching runtime context labels fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID, RuntimeContext2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when application is assigned if fetching listening runtimes labels fails",
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when application is assigned if fetching all listening runtimes fails",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return(nil, testErr)
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.RuntimeContext{fixRuntimeContextModel(), fixRuntimeContextModelWithRuntimeID(RuntimeID)}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when application is assigned if fetching runtime contexts in scenario fails",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return(nil, testErr)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when application is assigned if fetching runtimes in scenario fails",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixConfigurationChangedWebhookModel(WebhookID, ApplicationID, model.ApplicationWebhookReference), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctx, Tnt, []string{inputFormation.Name}).Return(nil, testErr)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success when generating notifications for application when application is assigned and no CONFIGURATION_CHANGED webhook is found",
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeID, RuntimeContextRuntimeID}) })).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil).Once()
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
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference), fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)}, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.AppWebhook, WebhookID))
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookID, RuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookID, RuntimeID), nil)
				conv.On("ToGraphQL", fixConfigurationChangedWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID, model.RuntimeWebhookReference)).Return(fixRuntimeWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				conv.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID), nil)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, mock.MatchedBy(func(ids []string) bool { return checkIfEqual(ids, []string{RuntimeContextRuntimeID, RuntimeID}) })).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID), fixRuntimeModel(RuntimeID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Times(3)
				return dataInputBuilder
			},
			ExpectedRequests: []*webhookclient.NotificationRequest{
				{
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
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
			Name: "error when generating notifications for application when application is assigned if fetching webhooks fails",
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(appWithLabels, appTemplateWithLabels, nil).Once()
				return dataInputBuilder
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when generating notifications for application when application is assigned if fetching application and/or app template labels fails",
			DataInputBuilder: func() *databuilderautomock.DataInputBuilder {
				dataInputBuilder := &databuilderautomock.DataInputBuilder{}
				dataInputBuilder.On("PrepareApplicationAndAppTemplateWithLabels", ctx, Tnt, ApplicationID).Return(nil, nil, testErr).Once()
				return dataInputBuilder
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
			dataInputBuilder := unusedDataInputBuilder()
			if testCase.DataInputBuilder != nil {
				dataInputBuilder = testCase.DataInputBuilder()
			}

			notificationSvc := formation.NewNotificationService(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookConverter, nil, dataInputBuilder)

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

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, applicationRepo, webhookRepo, webhookConverter, appTemplateRepo, labelRepo, dataInputBuilder)
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
				Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
					Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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
				Webhook: *fixRuntimeWebhookGQLModel(WebhookID, RuntimeID),
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

			notificationSvc := formation.NewNotificationService(nil, nil, nil, nil, nil, nil, nil, webhookClient, nil)

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

// helper func that checks if the elements of two slices are the same no matter their order
func checkIfEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	exists := make(map[string]bool)
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if !exists[value] {
			return false
		}
	}
	return true
}

func checkIfIDInSet(wh *model.Webhook, ids []string) bool {
	for _, id := range ids {
		if wh.ID == id {
			return true
		}
	}
	return false
}
