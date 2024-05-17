package runtime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	labelsWithNormalization       = map[string]interface{}{runtime.IsNormalizedLabel: "true"}
	protectedLabelPattern         = ".*_defaultEventing$|^consumer_subaccount_ids$"
	immutableLabelPattern         = "^xsappnameCMPClone$|^runtimeType$|^CMPSaaSAppName$"
	runtimeTypeLabelKey           = "runtimeType"
	regionLabelKey                = "region"
	regionLabelValue              = "test-region"
	kymaRuntimeTypeLabelValue     = "kyma"
	kymaApplicationNamespaceValue = "kyma.ns"
	testUUID                      = "b3ea1977-582e-4d61-ae12-b3a837a3858e"
	testScenario                  = "test-scenario"
)

func TestService_CreateWithMandatoryLabels(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	extSubaccountID := "extSubaccountID"
	subaccountID := "subaccountID"
	xsappNameCMPClone := "xsappnameCMPClone"
	xsappNameCMPCloneValue := "xsappnameCMPCloneValue"

	desc := "Lorem ipsum"
	labels := map[string]interface{}{
		"protected_defaultEventing": "true",
	}

	webhookInput := model.WebhookInput{
		Type: "type",
	}

	webhookMode := model.WebhookModeSync
	kymaWebhookInput := model.WebhookInput{
		Mode: &webhookMode,
		Type: model.WebhookTypeConfigurationChanged,
		Auth: &model.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		URLTemplate:    &urlTemplate,
		InputTemplate:  &inputTemplate,
		HeaderTemplate: &headerTemplate,
		OutputTemplate: &outputTemplate,
	}

	modelInput := func() model.RuntimeRegisterInput {
		return model.RuntimeRegisterInput{
			Name:        "foo.bar-not",
			Description: &desc,
			Labels:      labels,
			Webhooks: []*model.WebhookInput{{
				Type: "type",
			}},
		}
	}

	modelInputWithoutWebhooks := func() model.RuntimeRegisterInput {
		return model.RuntimeRegisterInput{
			Name:        "foo.bar-not",
			Description: &desc,
			Labels:      labels,
		}
	}

	modelInputWithSubaccountLabel := func() model.RuntimeRegisterInput {
		return model.RuntimeRegisterInput{
			Name:        "foo.bar-not",
			Description: &desc,
			Labels: map[string]interface{}{
				scenarioassignment.SubaccountIDKey: extSubaccountID,
			},
			Webhooks: []*model.WebhookInput{{
				Type: "type",
			}},
		}
	}

	modelInputWithScenariosLabel := func() model.RuntimeRegisterInput {
		return model.RuntimeRegisterInput{
			Name:        "foo.bar-not",
			Description: &desc,
			Labels: map[string]interface{}{
				model.ScenariosKey: []string{testScenario},
			},
			Webhooks: []*model.WebhookInput{{
				Type: "type",
			}},
		}
	}

	modelInputWithInvalidSubaccountLabel := func() model.RuntimeRegisterInput {
		return model.RuntimeRegisterInput{
			Name:        "foo.bar-not",
			Description: &desc,
			Labels: map[string]interface{}{
				scenarioassignment.SubaccountIDKey: 213,
			},
			Webhooks: []*model.WebhookInput{{
				Type: "type",
			}},
		}
	}

	labelsForDBMockWithSubaccount := map[string]interface{}{
		runtime.IsNormalizedLabel:          "true",
		scenarioassignment.SubaccountIDKey: extSubaccountID,
		runtimeTypeLabelKey:                kymaRuntimeTypeLabelValue,
		regionLabelKey:                     regionLabelValue,
	}

	labelsForDBMockWithMandatoryLabels := map[string]interface{}{
		runtime.IsNormalizedLabel: "true",
		xsappNameCMPClone:         xsappNameCMPCloneValue,
		runtimeTypeLabelKey:       kymaRuntimeTypeLabelValue,
		regionLabelKey:            "",
	}

	labelsForDBMockWithRuntimeType := map[string]interface{}{
		runtime.IsNormalizedLabel: "true",
		runtimeTypeLabelKey:       kymaRuntimeTypeLabelValue,
		regionLabelKey:            "",
	}

	modelRegionLabel := &model.Label{
		ID:         "id",
		Tenant:     &subaccountID,
		Key:        regionLabelKey,
		Value:      regionLabelValue,
		ObjectID:   subaccountID,
		ObjectType: model.TenantLabelableObject,
		Version:    0,
	}

	modelInputWithoutLabels := func() model.RuntimeRegisterInput {
		return model.RuntimeRegisterInput{
			Name:        "foo.bar-not",
			Description: &desc,
			Webhooks: []*model.WebhookInput{{
				Type: "type",
			}},
		}
	}

	var nilLabels map[string]interface{}

	runtimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput().Name && rtm.Description == modelInput().Description &&
			rtm.Status.Condition == model.RuntimeStatusConditionInitial
	})

	tnt := "tenant"
	externalTnt := "external-tnt"
	IntSysConsumer := consumer.Consumer{
		ConsumerID: "consumerID",
		Type:       consumer.IntegrationSystem,
		Flow:       oathkeeper.OAuth2Flow,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	ctxWithSubaccount := tenant.SaveToContext(ctx, subaccountID, extSubaccountID)
	ctxWithSubaccountAndIntSys := consumer.SaveToContext(ctxWithSubaccount, IntSysConsumer)
	ctxWithIntSysConsumer := consumer.SaveToContext(ctx, IntSysConsumer)

	ctxWithSubaccountMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		tenantCtx, err := tenant.LoadTenantPairFromContext(ctx)
		require.NoError(t, err)
		return subaccountID == tenantCtx.InternalID && extSubaccountID == tenantCtx.ExternalID
	})
	ctxWithGlobalaccountMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		tenantCtx, err := tenant.LoadTenantPairFromContext(ctx)
		require.NoError(t, err)
		return tnt == tenantCtx.InternalID
	})

	ga := &model.BusinessTenantMapping{
		ID:             tnt,
		Name:           "ga",
		ExternalTenant: externalTnt,
		Type:           "account",
		Provider:       "test",
		Status:         "Active",
	}

	subaccount := &model.BusinessTenantMapping{
		ID:             subaccountID,
		Name:           "sa",
		ExternalTenant: extSubaccountID,
		Parents:        []string{tnt},
		Type:           "subaccount",
		Provider:       "test",
		Status:         "Active",
	}

	subaccountInput := func() model.BusinessTenantMappingInput {
		return model.BusinessTenantMappingInput{
			ExternalTenant: extSubaccountID,
			Parents:        []string{tnt},
			Type:           "subaccount",
			Provider:       "lazilyWhileRuntimeCreation",
		}
	}

	testCases := []struct {
		Name                string
		RuntimeRepositoryFn func() *automock.RuntimeRepository
		TenantSvcFn         func() *automock.TenantService
		LabelServiceFn      func() *automock.LabelService
		WebhookServiceFn    func() *automock.WebhookService
		FormationServiceFn  func() *automock.FormationService
		Input               model.RuntimeRegisterInput
		MandatoryLabels     func() map[string]interface{}
		Context             context.Context
		ExpectedErr         error
	}{
		{
			Name: "Success",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithIntSysConsumer, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithIntSysConsumer, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithMandatoryLabels).Return(nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithIntSysConsumer, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", ctxWithIntSysConsumer, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", ctxWithIntSysConsumer, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				mandatoryLabels := make(map[string]interface{})
				mandatoryLabels[xsappNameCMPClone] = xsappNameCMPCloneValue
				mandatoryLabels[runtimeTypeLabelKey] = kymaRuntimeTypeLabelValue
				mandatoryLabels[regionLabelKey] = ""
				return mandatoryLabels
			},
			Context:     ctxWithIntSysConsumer,
			ExpectedErr: nil,
		},
		{
			Name: "Success without input webhooks",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithIntSysConsumer, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithIntSysConsumer, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithMandatoryLabels).Return(nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithIntSysConsumer, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", ctxWithIntSysConsumer, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			Input: modelInputWithoutWebhooks(),
			MandatoryLabels: func() map[string]interface{} {
				mandatoryLabels := make(map[string]interface{})
				mandatoryLabels[xsappNameCMPClone] = xsappNameCMPCloneValue
				mandatoryLabels[runtimeTypeLabelKey] = kymaRuntimeTypeLabelValue
				mandatoryLabels[regionLabelKey] = ""
				return mandatoryLabels
			},
			Context:     ctxWithIntSysConsumer,
			ExpectedErr: nil,
		},
		{
			Name: "Success with Subaccount label",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithSubaccountMatcher, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				svc.On("GetByKey", ctxWithSubaccountMatcher, subaccountID, model.TenantLabelableObject, subaccountID, regionLabelKey).Return(modelRegionLabel, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccountAndIntSys, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{tnt}}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", ctxWithSubaccountMatcher, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("AssignFormation", mock.Anything, tnt, runtimeID, graphql.FormationObjectTypeRuntime, model.Formation{Name: "test"}).Return(&model.Formation{Name: "test"}, nil).Once()
				svc.On("GetScenariosFromMatchingASAs", ctxWithGlobalaccountMatcher, runtimeID, graphql.FormationObjectTypeRuntime).Return([]string{"test"}, nil).Once()
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: nil,
		},
		{
			Name: "Success with Subaccount label when caller and label are the same",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithSubaccountMatcher, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				svc.On("GetByKey", ctxWithSubaccountMatcher, subaccountID, model.TenantLabelableObject, subaccountID, regionLabelKey).Return(modelRegionLabel, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				subaccountInput := subaccountInput()
				subaccountInput.Parents = []string{subaccountID}
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccountAndIntSys, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{tnt}}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("AssignFormation", mock.Anything, tnt, runtimeID, graphql.FormationObjectTypeRuntime, model.Formation{Name: "test"}).Return(&model.Formation{Name: "test"}, nil).Once()
				svc.On("GetScenariosFromMatchingASAs", ctxWithGlobalaccountMatcher, runtimeID, graphql.FormationObjectTypeRuntime).Return([]string{"test"}, nil).Once()
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: nil,
		},
		{
			Name: "Success with Subaccount label and no scenarios from ASAs in parent",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithSubaccountMatcher, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				svc.On("GetByKey", ctxWithSubaccountMatcher, subaccountID, model.TenantLabelableObject, subaccountID, regionLabelKey).Return(modelRegionLabel, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccountAndIntSys, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{tnt}}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("GetScenariosFromMatchingASAs", ctxWithGlobalaccountMatcher, runtimeID, graphql.FormationObjectTypeRuntime).Return([]string{}, nil).Once()
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: nil,
		},
		{
			Name: "Success when labels are empty",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithIntSysConsumer, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithIntSysConsumer, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithRuntimeType).Return(nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithIntSysConsumer, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			Input: modelInputWithoutLabels(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithIntSysConsumer,
			ExpectedErr: nil,
		},
		{
			Name:  "Returns error when subaccount label conversion fail",
			Input: modelInputWithInvalidSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: errors.New("while converting global_subaccount_id label"),
		},
		{
			Name: "Returns error when subaccount get from DB fail",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(nil, testErr).Once()
				return tenantSvc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when runtime creation failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(testErr).Once()
				return repo
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when subaccount in the label is not child of the caller",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{"anotherParent"}}, nil).Once()
				return tenantSvc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: apperrors.NewInvalidOperationError(fmt.Sprintf("Tenant provided in %s label should be child of the caller tenant", scenarioassignment.SubaccountIDKey)),
		},
		{
			Name: "Return error when get calling tenant from DB fail",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithSubaccountMatcher, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				svc.On("GetByKey", ctxWithSubaccountMatcher, subaccountID, model.TenantLabelableObject, subaccountID, regionLabelKey).Return(modelRegionLabel, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccountMatcher, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{tnt}}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(nil, testErr).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when webhook creation failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithIntSysConsumer, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithIntSysConsumer, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithMandatoryLabels).Return(nil).Once()
				return svc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("", testErr).Once()
				return webhookSvc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				mandatoryLabels := make(map[string]interface{})
				mandatoryLabels[xsappNameCMPClone] = xsappNameCMPCloneValue
				mandatoryLabels[runtimeTypeLabelKey] = kymaRuntimeTypeLabelValue
				return mandatoryLabels
			},
			Context:     ctxWithIntSysConsumer,
			ExpectedErr: testErr,
		},
		{
			Name: "Return error when getting scenarios from ASA failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithSubaccountMatcher, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				svc.On("GetByKey", ctxWithSubaccountMatcher, subaccountID, model.TenantLabelableObject, subaccountID, regionLabelKey).Return(modelRegionLabel, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccountMatcher, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{tnt}}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("GetScenariosFromMatchingASAs", ctxWithGlobalaccountMatcher, runtimeID, graphql.FormationObjectTypeRuntime).Return(nil, testErr).Once()
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: testErr,
		},
		{
			Name:  "Return error when when there is scenario label in the input",
			Input: modelInputWithScenariosLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: errors.Errorf("label with key %s cannot be set explicitly", model.ScenariosKey),
		},
		{
			Name: "Returns error when getting region label failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctxWithSubaccountMatcher, subaccountID, model.TenantLabelableObject, subaccountID, regionLabelKey).Return(nil, testErr).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccountAndIntSys, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{tnt}}, nil).Once()
				return tenantSvc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when label upserting failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithIntSysConsumer, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithIntSysConsumer, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithRuntimeType).Return(testErr).Once()
				return svc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithIntSysConsumer,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when can't assign scenarios to parent",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithSubaccountMatcher, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				svc.On("GetByKey", ctxWithSubaccountMatcher, subaccountID, model.TenantLabelableObject, subaccountID, regionLabelKey).Return(modelRegionLabel, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccountMatcher, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parents: []string{tnt}}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				svc.On("AssignFormation", mock.Anything, tnt, runtimeID, graphql.FormationObjectTypeRuntime, model.Formation{Name: "test"}).Return(nil, testErr).Once()
				svc.On("GetScenariosFromMatchingASAs", ctxWithGlobalaccountMatcher, runtimeID, graphql.FormationObjectTypeRuntime).Return([]string{"test"}, nil).Once()
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccountAndIntSys,
			ExpectedErr: testErr,
		},
		{
			Name: "Successfully added runtime type label when the consumer type is integration system",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithIntSysConsumer, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertMultipleLabels", ctxWithIntSysConsumer, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithRuntimeType).Return(nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithIntSysConsumer, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			WebhookServiceFn: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil).Once()
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, kymaWebhookInput, model.RuntimeWebhookReference).Return("kymaWebhookID", nil).Once()
				return webhookSvc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithIntSysConsumer,
			ExpectedErr: nil,
		},
		{
			Name:  "Returns error when there is no consumer in the context",
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: errors.New("while loading consumer: Internal Server Error: cannot read consumer from context"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := &automock.RuntimeRepository{}
			if testCase.RuntimeRepositoryFn != nil {
				repo = testCase.RuntimeRepositoryFn()
			}
			labelSvc := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelSvc = testCase.LabelServiceFn()
			}
			tenantSvc := unusedTenantService()
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}
			formationSvc := unusedFormationService()
			if testCase.FormationServiceFn != nil {
				formationSvc = testCase.FormationServiceFn()
			}
			webhookSvc := unusedWebhookService()
			if testCase.WebhookServiceFn != nil {
				webhookSvc = testCase.WebhookServiceFn()
			}
			mandatoryLabels := testCase.MandatoryLabels()
			svc := runtime.NewService(repo, nil, labelSvc, nil, formationSvc, tenantSvc, webhookSvc, nil, protectedLabelPattern, immutableLabelPattern, runtimeTypeLabelKey, kymaRuntimeTypeLabelValue, kymaApplicationNamespaceValue, string(webhookMode), webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			err := svc.CreateWithMandatoryLabels(testCase.Context, testCase.Input, runtimeID, mandatoryLabels)

			// then
			if err == nil {
				require.Nil(t, testCase.ExpectedErr)
			} else {
				require.NotNil(t, testCase.ExpectedErr)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, repo, labelSvc, tenantSvc, formationSvc, webhookSvc)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		uuidSvc := &automock.UidService{}
		uuidSvc.On("Generate").Return(testUUID).Once()

		svc := runtime.NewService(nil, nil, nil, uuidSvc, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", string(webhookMode), webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		_, err := svc.Create(context.TODO(), model.RuntimeRegisterInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
		uuidSvc.AssertExpectations(t)
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	desc := "Lorem ipsum"

	labelsDBMock := map[string]interface{}{
		"label1":                  "val1",
		runtime.IsNormalizedLabel: "true",
	}
	labels := map[string]interface{}{
		"label1": "val1",
	}
	protectedLabels := map[string]interface{}{
		"protected_defaultEventing": "true",
		"label1":                    "val1",
	}
	modelInput := model.RuntimeUpdateInput{
		Name:   "bar",
		Labels: labels,
	}

	modelInputWithProtectedLabels := model.RuntimeUpdateInput{
		Name:   "bar",
		Labels: protectedLabels,
	}

	modelInputWithScenariosLabel := model.RuntimeUpdateInput{
		Name: "bar",
		Labels: map[string]interface{}{
			model.ScenariosKey: []string{testScenario},
		},
	}

	inputRuntimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name
	})

	inputProtectedRuntimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name
	})

	runtimeModel := &model.Runtime{
		ID:          runtimeID,
		Name:        "Foo",
		Description: &desc,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                string
		RuntimeRepositoryFn func() *automock.RuntimeRepository
		LabelRepositoryFn   func() *automock.LabelRepository
		LabelServiceFn      func() *automock.LabelService
		Input               model.RuntimeUpdateInput
		InputID             string
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, tnt, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				repo := &automock.LabelService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, modelInput.Labels).Return(nil).Once()
				return repo
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when updating with protected labels",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, tnt, inputProtectedRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				repo := &automock.LabelService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labelsDBMock).Return(nil).Once()
				return repo
			},
			InputID:            runtimeID,
			Input:              modelInputWithProtectedLabels,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when labels are nil",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, tnt, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				repo := &automock.LabelService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labelsWithNormalization).Return(nil).Once()
				return repo
			},
			InputID: runtimeID,
			Input: model.RuntimeUpdateInput{
				Name: "bar",
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, tnt, inputRuntimeModel).Return(testErr).Once()
				return repo
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when there is scenarios label in the input",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				return repo
			},
			InputID:            runtimeID,
			Input:              modelInputWithScenariosLabel,
			ExpectedErrMessage: errors.Errorf("label with key %s cannot be set explicitly", model.ScenariosKey).Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label deletion failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, tnt, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(testErr).Once()
				return repo
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when upserting labels failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, tnt, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				repo := &automock.LabelService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, modelInput.Labels).Return(testErr).Once()
				return repo
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := &automock.RuntimeRepository{}
			if testCase.RuntimeRepositoryFn != nil {
				repo = testCase.RuntimeRepositoryFn()
			}
			labelRepo := &automock.LabelRepository{}
			if testCase.LabelRepositoryFn != nil {
				labelRepo = testCase.LabelRepositoryFn()
			}
			labelSvc := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelSvc = testCase.LabelServiceFn()
			}
			svc := runtime.NewService(repo, labelRepo, labelSvc, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelRepo, labelSvc)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		err := svc.Update(context.TODO(), "id", model.RuntimeUpdateInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	rtmCtxID := "rtmCtx"

	desc := "Lorem ipsum"

	runtimeModel := &model.Runtime{
		ID:          id,
		Name:        "Foo",
		Description: &desc,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	runtimeContext := &model.RuntimeContext{
		ID:        rtmCtxID,
		RuntimeID: id,
		Key:       "test",
		Value:     "test",
	}
	runtimeContexts := []*model.RuntimeContext{runtimeContext}

	formations := []*model.Formation{{Name: "scenario1"}, {Name: "scenario2"}}

	testCases := []struct {
		Name                string
		RepositoryFn        func() *automock.RuntimeRepository
		RuntimeContextSvcFn func() *automock.RuntimeContextService
		LabelRepoFn         func() *automock.LabelRepository
		FormationServiceFn  func() *automock.FormationService
		InputID             string
		ExpectedErrMessage  string
	}{
		{
			Name: "Success for runtime with formations",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, id).Return(nil).Once()
				return repo
			},
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(runtimeContexts, nil).Once()
				runtimeContextSvc.On("Delete", ctx, rtmCtxID).Return(nil).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				engine := &automock.FormationService{}
				engine.On("ListFormationsForObject", ctx, id).Return(formations, nil).Once()

				engine.On("UnassignFormation", ctx, tnt, id, graphql.FormationObjectTypeRuntime, model.Formation{Name: formations[0].Name}, true).Return(&model.Formation{Name: "scenario1"}, nil)
				engine.On("UnassignFormation", ctx, tnt, id, graphql.FormationObjectTypeRuntime, model.Formation{Name: formations[1].Name}, true).Return(&model.Formation{Name: "scenario2"}, nil)
				return engine
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success for runtime without formations",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, id).Return(nil).Once()
				return repo
			},
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(runtimeContexts, nil).Once()
				runtimeContextSvc.On("Delete", ctx, rtmCtxID).Return(nil).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, id).Return(nil, nil).Once()
				return formationSvc
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error while listing runtime contexts",
			RepositoryFn: func() *automock.RuntimeRepository {
				return &automock.RuntimeRepository{}
			},
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(nil, testErr).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: unusedFormationService,
			InputID:            id,
			ExpectedErrMessage: "while listing runtimeContexts for runtime",
		},
		{
			Name: "Returns error while deleting runtime context",
			RepositoryFn: func() *automock.RuntimeRepository {
				return &automock.RuntimeRepository{}
			},
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(runtimeContexts, nil).Once()
				runtimeContextSvc.On("Delete", ctx, rtmCtxID).Return(testErr).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: unusedFormationService,
			InputID:            id,
			ExpectedErrMessage: "while deleting runtimeContext",
		},
		{
			Name: "Returns error when runtime deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, id).Return(nil).Once()
				return repo
			},
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(runtimeContexts, nil).Once()
				runtimeContextSvc.On("Delete", ctx, rtmCtxID).Return(nil).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationService := &automock.FormationService{}
				formationService.On("ListFormationsForObject", ctx, id).Return(nil, nil).Once()
				return formationService
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name:         "Returns error when unassign formation fails",
			RepositoryFn: unusedRuntimeRepository,
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(runtimeContexts, nil).Once()
				runtimeContextSvc.On("Delete", ctx, rtmCtxID).Return(nil).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				engine := &automock.FormationService{}
				engine.On("ListFormationsForObject", ctx, id).Return(formations, nil).Once()

				engine.On("UnassignFormation", ctx, tnt, id, graphql.FormationObjectTypeRuntime, model.Formation{Name: formations[0].Name}, true).Return(nil, testErr)
				return engine
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:         "Returns error when listing current runtime formation fails",
			RepositoryFn: unusedRuntimeRepository,
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(runtimeContexts, nil).Once()
				runtimeContextSvc.On("Delete", ctx, rtmCtxID).Return(nil).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, id).Return(nil, testErr).Once()
				return formationSvc
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, runtimeModel.ID).Return(testErr).Once()
				return repo
			},
			RuntimeContextSvcFn: func() *automock.RuntimeContextService {
				runtimeContextSvc := &automock.RuntimeContextService{}
				runtimeContextSvc.On("ListAllForRuntime", ctx, id).Return(runtimeContexts, nil).Once()
				runtimeContextSvc.On("Delete", ctx, rtmCtxID).Return(nil).Once()
				return runtimeContextSvc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("ListFormationsForObject", ctx, id).Return(nil, nil).Once()
				return formationSvc
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := unusedLabelRepository()
			if testCase.LabelRepoFn != nil {
				labelRepo = testCase.LabelRepoFn()
			}
			engine := testCase.FormationServiceFn()
			rtmCtxSvc := testCase.RuntimeContextSvcFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil, engine, nil, nil, rtmCtxSvc, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelRepo, engine, rtmCtxSvc)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		err := svc.Delete(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := runtimeID
	desc := "Lorem ipsum"
	tnt := "tenant"
	externalTnt := "external-tnt"

	runtimeModel := &model.Runtime{
		ID:          runtimeID,
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputID            string
		ExpectedRuntime    *model.Runtime
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			rtm, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntime, rtm)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		_, err := svc.Get(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_GetByTokenIssuer(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	desc := "Lorem ipsum"
	tokenIssuer := "https://dex.domain.local"
	filter := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("runtime_consoleUrl", `"https://console.domain.local"`)}

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Input              model.RuntimeRegisterInput
		InputID            string
		ExpectedRuntime    *model.Runtime
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filter).Return(runtimeModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filter).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			rtm, err := svc.GetByTokenIssuer(ctx, tokenIssuer)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntime, rtm)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Exist(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")

	rtmID := "id"

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.RuntimeRepository
		InputRuntimeID string
		ExpectedValue  bool
		ExpectedError  error
	}{
		{
			Name: "Runtime exits",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, rtmID).Return(true, nil)
				return repo
			},
			InputRuntimeID: rtmID,
			ExpectedValue:  true,
			ExpectedError:  nil,
		},
		{
			Name: "Runtime not exits",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, rtmID).Return(false, nil)
				return repo
			},
			InputRuntimeID: rtmID,
			ExpectedValue:  false,
			ExpectedError:  nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, rtmID).Return(false, testError)
				return repo
			},
			InputRuntimeID: rtmID,
			ExpectedValue:  false,
			ExpectedError:  testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			rtmRepo := testCase.RepositoryFn()
			svc := runtime.NewService(rtmRepo, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			value, err := svc.Exist(ctx, testCase.InputRuntimeID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, testCase.ExpectedValue, value)
			rtmRepo.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		_, err := svc.Exist(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_List(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	modelRuntimes := []*model.Runtime{
		fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem Ipsum", "test.ns.foo"),
		fixModelRuntime(t, "bar", "tenant-bar", "Bar", "Lorem Ipsum", "test.ns.bar"),
	}
	runtimePage := &model.RuntimePage{
		Data:       modelRuntimes,
		TotalCount: len(modelRuntimes),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"
	filter := []*labelfilter.LabelFilter{{Key: ""}}

	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      int
		InputCursor        string
		ExpectedResult     *model.RuntimePage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("List", ctx, tnt, filter, first, after).Return(runtimePage, nil).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     runtimePage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime listing failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("List", ctx, tnt, filter, first, after).Return(nil, testErr).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "Returns error when pageSize is less than 1",
			RepositoryFn:       unusedRuntimeRepository,
			InputLabelFilters:  filter,
			InputPageSize:      0,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name:               "Returns error when pageSize is bigger than 200",
			RepositoryFn:       unusedRuntimeRepository,
			InputLabelFilters:  filter,
			InputPageSize:      201,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			rtm, err := svc.List(ctx, testCase.InputLabelFilters, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, rtm)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		_, err := svc.List(context.TODO(), nil, 1, "")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_GetLabel(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputRuntimeID     string
		InputLabel         *model.LabelInput
		ExpectedLabel      *model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(modelLabel, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedLabel:      modelLabel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when label receiving failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil, testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedLabel:      nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when exists function for runtime failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn:  unusedLabelRepository,
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime doesn't exist",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()

				return repo
			},
			LabelRepositoryFn:  unusedLabelRepository,
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			l, err := svc.GetLabel(ctx, testCase.InputRuntimeID, testCase.InputLabel.Key)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedLabel)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		_, err := svc.GetLabel(context.TODO(), "id", "key")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_ListLabels(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	protectedModelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c12",
		Key:        "protected_defaultEventing",
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	labels := map[string]*model.Label{"protected_defaultEventing": protectedModelLabel, "first": modelLabel, "second": modelLabel}
	expectedLabelWithoutProtected := map[string]*model.Label{"first": modelLabel, "second": modelLabel}
	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputRuntimeID     string
		InputLabel         *model.LabelInput
		ExpectedOutput     map[string]*model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labels, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedOutput:     expectedLabelWithoutProtected,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when labels receiving failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedOutput:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime exists function failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn:  unusedLabelRepository,
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime does not exists",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()

				return repo
			},
			LabelRepositoryFn:  unusedLabelRepository,
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			l, err := svc.ListLabels(ctx, testCase.InputRuntimeID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, l)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "", "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		_, err := svc.ListLabels(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_SetLabel(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"

	labelKey := "key"
	protectedLabelKey := "protected_defaultEventing"

	modelLabelInput := model.LabelInput{
		Key:        labelKey,
		Value:      []string{"value1"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelProtectedLabelInput := model.LabelInput{
		Key:        protectedLabelKey,
		Value:      []string{"value1"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name                string
		RuntimeRepositoryFn func() *automock.RuntimeRepository
		LabelServiceFn      func() *automock.LabelService
		InputRuntimeID      string
		InputLabel          *model.LabelInput
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, &modelLabelInput).Return(nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if runtime exists failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime doesn't exists",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
		{
			Name: "Returns error when upsert label fails",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, tnt, &modelLabelInput).Return(testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns an error when trying to set protected label",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelProtectedLabelInput,
			ExpectedErrMessage: "could not set unmodifiable label with key protected_defaultEventing",
		},
		{
			Name: "Returns an error when trying to set scenarios label",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &model.LabelInput{Key: model.ScenariosKey, ObjectID: runtimeID},
			ExpectedErrMessage: fmt.Sprintf("label with key %s cannot be set explicitly", model.ScenariosKey),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := &automock.RuntimeRepository{}
			if testCase.RuntimeRepositoryFn != nil {
				repo = testCase.RuntimeRepositoryFn()
			}
			labelSvc := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelSvc = testCase.LabelServiceFn()
			}
			svc := runtime.NewService(repo, nil, labelSvc, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			err := svc.SetLabel(ctx, testCase.InputLabel)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelSvc)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		err := svc.SetLabel(context.TODO(), &model.LabelInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_DeleteLabel(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"

	labelKey := "key"
	protectedLabelKey := "protected_defaultEventing"

	testCases := []struct {
		Name                string
		RuntimeRepositoryFn func() *automock.RuntimeRepository
		LabelRepositoryFn   func() *automock.LabelRepository
		InputRuntimeID      string
		InputKey            string
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if runtime exists failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime does not exists",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
		{
			Name: "Returns error when runtime label delete failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns an error when trying to delete protected label",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           protectedLabelKey,
			ExpectedErrMessage: "could not delete unmodifiable label with key protected_defaultEventing",
		},
		{
			Name: "Returns an error when trying to delete scenarios label",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: fmt.Sprintf("label with key %s cannot be deleted explicitly", model.ScenariosKey),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := &automock.RuntimeRepository{}
			if testCase.RuntimeRepositoryFn != nil {
				repo = testCase.RuntimeRepositoryFn()
			}
			labelRepo := &automock.LabelRepository{}
			if testCase.LabelRepositoryFn != nil {
				labelRepo = testCase.LabelRepositoryFn()
			}
			svc := runtime.NewService(repo, labelRepo, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			err := svc.DeleteLabel(ctx, testCase.InputRuntimeID, testCase.InputKey)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelRepo)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)
		// WHEN
		err := svc.DeleteLabel(context.TODO(), "id", "key")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_GetByFiltersGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	filters := []*labelfilter.LabelFilter{
		{Key: "test-key", Query: str.Ptr("test-filter")},
	}
	testRuntime := &model.Runtime{
		ID:   "test-id",
		Name: "test-runtime",
	}
	ctx := context.TODO()

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filters).Return(testRuntime, nil).Once()
				return repo
			},

			ExpectedErrMessage: "",
		},
		{
			Name: "Fails on repository error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filters).Return(nil, testErr).Once()
				return repo
			},

			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepository := &automock.LabelRepository{}
			labelService := unusedLabelService()
			formationService := &automock.FormationService{}
			uidSvc := &automock.UidService{}
			svc := runtime.NewService(repo, labelRepository, labelService, uidSvc, formationService, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			actualRuntime, err := svc.GetByFiltersGlobal(ctx, filters)
			// then
			if testCase.ExpectedErrMessage == "" {
				require.Equal(t, testRuntime, actualRuntime)
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelService, labelRepository, formationService, uidSvc)
		})
	}
}

func TestService_GetByFilters(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	testErr := errors.New("Test error")
	filters := []*labelfilter.LabelFilter{
		{Key: "test-key", Query: str.Ptr("test-filter")},
	}
	modelRuntime := fixModelRuntime(t, "foo", tnt, "Foo", "Lorem Ipsum", "test.ns")
	ctx := tenant.SaveToContext(context.TODO(), tnt, tnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Context            context.Context
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFilters", contextThatHasTenant(tnt), tnt, filters).Return(modelRuntime, nil).Once()
				return repo
			},
			Context:            ctx,
			ExpectedErrMessage: "",
		},
		{
			Name: "Fails on repository error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFilters", contextThatHasTenant(tnt), tnt, filters).Return(nil, testErr).Once()
				return repo
			},
			Context:            ctx,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "Fails when no tenant in the context",
			RepositoryFn:       unusedRuntimeRepository,
			Context:            context.TODO(),
			ExpectedErrMessage: "while loading tenant from context",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepository := &automock.LabelRepository{}
			labelService := unusedLabelService()
			formationService := &automock.FormationService{}
			uidSvc := &automock.UidService{}
			svc := runtime.NewService(repo, labelRepository, labelService, uidSvc, formationService, nil, nil, nil, ".*_defaultEventing$", immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			actualRuntime, err := svc.GetByFilters(testCase.Context, filters)
			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				require.Equal(t, modelRuntime, actualRuntime)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelService, labelRepository, formationService, uidSvc)
		})
	}
}

func TestService_ListByFiltersGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	filters := []*labelfilter.LabelFilter{
		{Key: "test-key", Query: str.Ptr("test-filter")},
	}
	modelRuntimes := []*model.Runtime{
		fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem Ipsum", "test.ns.foo"),
		fixModelRuntime(t, "bar", "tenant-bar", "Bar", "Lorem Ipsum", "test.ns.bar"),
	}
	ctx := context.TODO()

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByFiltersGlobal", ctx, filters).Return(modelRuntimes, nil).Once()
				return repo
			},

			ExpectedErrMessage: "",
		},
		{
			Name: "Fails on repository error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByFiltersGlobal", ctx, filters).Return(nil, testErr).Once()
				return repo
			},

			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepository := &automock.LabelRepository{}
			labelService := unusedLabelService()
			formationService := &automock.FormationService{}
			uidSvc := &automock.UidService{}
			svc := runtime.NewService(repo, labelRepository, labelService, uidSvc, formationService, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			actualRuntimes, err := svc.ListByFiltersGlobal(ctx, filters)
			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				require.Equal(t, modelRuntimes, actualRuntimes)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelService, labelRepository, formationService, uidSvc)
		})
	}
}

func TestService_ListByFilters(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	testErr := errors.New("Test error")
	filters := []*labelfilter.LabelFilter{
		{Key: "test-key", Query: str.Ptr("test-filter")},
	}
	modelRuntimes := []*model.Runtime{
		fixModelRuntime(t, "foo", tnt, "Foo", "Lorem Ipsum", "test.ns.foo"),
		fixModelRuntime(t, "bar", tnt, "Bar", "Lorem Ipsum", "test.ns.bar"),
	}
	ctx := tenant.SaveToContext(context.TODO(), tnt, tnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Context            context.Context
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", contextThatHasTenant(tnt), tnt, filters).Return(modelRuntimes, nil).Once()
				return repo
			},
			Context:            ctx,
			ExpectedErrMessage: "",
		},
		{
			Name: "Fails on repository error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListAll", contextThatHasTenant(tnt), tnt, filters).Return(nil, testErr).Once()
				return repo
			},
			Context:            ctx,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "Fails when no tenant in the context",
			RepositoryFn:       unusedRuntimeRepository,
			Context:            context.TODO(),
			ExpectedErrMessage: "while loading tenant from context",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepository := &automock.LabelRepository{}
			labelService := unusedLabelService()
			formationService := &automock.FormationService{}
			uidSvc := &automock.UidService{}
			svc := runtime.NewService(repo, labelRepository, labelService, uidSvc, formationService, nil, nil, nil, ".*_defaultEventing$", immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			actualRuntimes, err := svc.ListByFilters(testCase.Context, filters)
			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				require.Equal(t, modelRuntimes, actualRuntimes)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, labelService, labelRepository, formationService, uidSvc)
		})
	}
}

func TestService_UnsafeExtractModifiableLabels(t *testing.T) {
	testCases := []struct {
		Name           string
		InputLabels    map[string]interface{}
		ExpectedLabels map[string]interface{}
		ExpectedErr    error
	}{
		{
			Name:           "Success without protected and immutable labels",
			InputLabels:    map[string]interface{}{"test1": "test1", "test2": "test2"},
			ExpectedLabels: map[string]interface{}{"test1": "test1", "test2": "test2"},
			ExpectedErr:    nil,
		},
		{
			Name:           "Success with protected labels",
			InputLabels:    map[string]interface{}{"test_defaultEventing": "protected", "test2": "test2"},
			ExpectedLabels: map[string]interface{}{"test2": "test2"},
			ExpectedErr:    nil,
		},
		{
			Name:           "Success with immutable labels",
			InputLabels:    map[string]interface{}{runtimeTypeLabelKey: "immutable", "test2": "test2"},
			ExpectedLabels: map[string]interface{}{"test2": "test2"},
			ExpectedErr:    nil,
		},
		{
			Name:           "Success with protected and immutable labels",
			InputLabels:    map[string]interface{}{runtimeTypeLabelKey: "test1", "test_defaultEventing": "test2", "test3": "test3"},
			ExpectedLabels: map[string]interface{}{"test3": "test3"},
			ExpectedErr:    nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern, "", "", "", webhookMode, webhookType, urlTemplate, inputTemplate, headerTemplate, outputTemplate)

			// WHEN
			extractedLabels, err := svc.UnsafeExtractModifiableLabels(testCase.InputLabels)
			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Equal(t, nil, extractedLabels)
			} else {
				require.NoError(t, err)
				require.Equal(t, extractedLabels, testCase.ExpectedLabels)
			}
		})
	}
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}

func unusedFormationService() *automock.FormationService {
	return &automock.FormationService{}
}

func unusedRuntimeRepository() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func unusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func unusedTenantService() *automock.TenantService {
	return &automock.TenantService{}
}

func unusedWebhookService() *automock.WebhookService {
	return &automock.WebhookService{}
}

func unusedLabelRepository() *automock.LabelRepository {
	return &automock.LabelRepository{}
}
