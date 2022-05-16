package runtime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/rtmtest"

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
	labelsWithNormalization = map[string]interface{}{runtime.IsNormalizedLabel: "true"}
	protectedLabelPattern   = ".*_defaultEventing$|^consumer_subaccount_ids$"
	immutableLabelPattern   = "^xsappnameCMPClone$"
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
		model.ScenariosKey:          []interface{}{"DEFAULT"},
		"protected_defaultEventing": "true",
		"consumer_subaccount_ids":   []string{"subaccountID-1", "subaccountID-2"},
	}
	labelsForDBMock := map[string]interface{}{
		model.ScenariosKey:        []interface{}{"DEFAULT"},
		runtime.IsNormalizedLabel: "true",
	}

	webhookInput := model.WebhookInput{
		Type: "type",
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
				model.ScenariosKey:                 []interface{}{"DEFAULT"},
				scenarioassignment.SubaccountIDKey: extSubaccountID,
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
				model.ScenariosKey:                 []interface{}{"DEFAULT"},
				scenarioassignment.SubaccountIDKey: 213,
			},
			Webhooks: []*model.WebhookInput{{
				Type: "type",
			}},
		}
	}

	labelsForDBMockWithSubaccount := map[string]interface{}{
		model.ScenariosKey:                 []interface{}{"DEFAULT"},
		runtime.IsNormalizedLabel:          "true",
		scenarioassignment.SubaccountIDKey: extSubaccountID,
	}

	labelsForDBMockWithoutNormalization := map[string]interface{}{
		model.ScenariosKey:                 []interface{}{"DEFAULT"},
		scenarioassignment.SubaccountIDKey: extSubaccountID,
	}

	labelsForDBMockWithXsappName := map[string]interface{}{
		model.ScenariosKey:        []interface{}{"DEFAULT"},
		runtime.IsNormalizedLabel: "true",
		xsappNameCMPClone:         xsappNameCMPCloneValue,
	}

	parentScenarios := map[string]interface{}{
		model.ScenariosKey: []interface{}{"test"},
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
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	ctxWithSubaccount := tenant.SaveToContext(ctx, subaccountID, extSubaccountID)

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
		Parent:         tnt,
		Type:           "subaccount",
		Provider:       "test",
		Status:         "Active",
	}

	subaccountInput := func() model.BusinessTenantMappingInput {
		return model.BusinessTenantMappingInput{
			ExternalTenant: extSubaccountID,
			Parent:         tnt,
			Type:           "subaccount",
			Provider:       "lazilyWhileRuntimeCreation",
		}
	}

	testCases := []struct {
		Name                 string
		RuntimeRepositoryFn  func() *automock.RuntimeRepository
		ScenariosServiceFn   func() *automock.ScenariosService
		TenantSvcFn          func() *automock.TenantService
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		UIDServiceFn         func() *automock.UidService
		WebhookService       func() *automock.WebhookService
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		Input                model.RuntimeRegisterInput
		MandatoryLabels      func() map[string]interface{}
		Context              context.Context
		ExpectedErr          error
	}{
		{
			Name: "Success",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctx, tnt, &labels).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithXsappName).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctx, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				mandatoryLabels := make(map[string]interface{})
				mandatoryLabels[xsappNameCMPClone] = xsappNameCMPCloneValue
				return mandatoryLabels
			},
			Context:     ctx,
			ExpectedErr: nil,
		},
		{
			Name: "Success without webhooks",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctx, tnt, &labels).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithXsappName).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctx, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			UIDServiceFn:   rtmtest.UnusedUUIDService(),
			WebhookService: rtmtest.UnusedWebhookService(),
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInputWithoutWebhooks(),
			MandatoryLabels: func() map[string]interface{} {
				mandatoryLabels := make(map[string]interface{})
				mandatoryLabels[xsappNameCMPClone] = xsappNameCMPCloneValue
				return mandatoryLabels
			},
			Context:     ctx,
			ExpectedErr: nil,
		},
		{
			Name: "Success with Subaccount label",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccount, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctxWithSubaccount, subaccountID, &labelsForDBMockWithoutNormalization).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctxWithSubaccount, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				repo.On("UpsertMultipleLabels", ctxWithGlobalaccountMatcher, tnt, model.RuntimeLabelableObject, runtimeID, parentScenarios).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parent: tnt}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctxWithGlobalaccountMatcher, map[string]interface{}{}, runtimeID).Return([]interface{}{"test"}, nil)
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: nil,
		},
		{
			Name: "Success with Subaccount label when caller and label are the same",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccountMatcher, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctxWithSubaccountMatcher, subaccountID, &labelsForDBMockWithoutNormalization).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctxWithSubaccountMatcher, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				repo.On("UpsertMultipleLabels", ctxWithGlobalaccountMatcher, tnt, model.RuntimeLabelableObject, runtimeID, parentScenarios).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				subaccountInput := subaccountInput()
				subaccountInput.Parent = subaccountID
				tenantSvc.On("GetTenantByExternalID", ctxWithSubaccount, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parent: tnt}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctxWithGlobalaccountMatcher, map[string]interface{}{}, runtimeID).Return([]interface{}{"test"}, nil)
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctxWithSubaccount,
			ExpectedErr: nil,
		},
		{
			Name: "Success with Subaccount label and no scenarios from ASAs in parent",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccount, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctxWithSubaccount, subaccountID, &labelsForDBMockWithoutNormalization).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctxWithSubaccount, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parent: tnt}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctxWithGlobalaccountMatcher, map[string]interface{}{}, runtimeID).Return([]interface{}{}, nil)
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: nil,
		},
		{
			Name: "Success when labels are empty",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, tnt, &nilLabels).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, runtimeID, labelsWithNormalization).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctx, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInputWithoutLabels(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when subaccount label conversion fail",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				return &automock.ScenariosService{}
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			UIDServiceFn:   rtmtest.UnusedUUIDService(),
			WebhookService: rtmtest.UnusedWebhookService(),
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInputWithInvalidSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: errors.New("while converting global_subaccount_id label"),
		},
		{
			Name: "Returns error when subaccount get from DB fail",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				return &automock.ScenariosService{}
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(nil, testErr).Once()
				return tenantSvc
			},
			UIDServiceFn:   rtmtest.UnusedUUIDService(),
			WebhookService: rtmtest.UnusedWebhookService(),
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
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
				repo.On("Create", ctx, tnt, runtimeModel).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				return repo
			},
			UIDServiceFn:   rtmtest.UnusedUUIDService(),
			WebhookService: rtmtest.UnusedWebhookService(),
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when subaccount in the label is not child of the caller",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				return &automock.ScenariosService{}
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parent: "anotherParent"}, nil).Once()
				return tenantSvc
			},
			UIDServiceFn:   rtmtest.UnusedUUIDService(),
			WebhookService: rtmtest.UnusedWebhookService(),
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
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
				repo.On("Create", ctxWithSubaccount, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctxWithSubaccount, subaccountID, &labelsForDBMockWithoutNormalization).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctxWithSubaccount, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parent: tnt}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(nil, testErr).Once()
				return tenantSvc
			},
			UIDServiceFn:   rtmtest.UnusedUUIDService(),
			WebhookService: rtmtest.UnusedWebhookService(),
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when webhook creation failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctx, tnt, &labels).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithXsappName).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctx, tnt).Return(ga, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("", testErr)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				mandatoryLabels := make(map[string]interface{})
				mandatoryLabels[xsappNameCMPClone] = xsappNameCMPCloneValue
				return mandatoryLabels
			},
			Context:     ctx,
			ExpectedErr: testErr,
		},
		{
			Name: "Return error when merge of scenarios and assignments failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccount, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctxWithSubaccount, subaccountID, &labelsForDBMockWithoutNormalization).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctxWithSubaccount, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parent: tnt}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctxWithGlobalaccountMatcher, map[string]interface{}{}, runtimeID).Return(nil, testErr)
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when label upserting failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, tnt, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctx, tnt, &labels).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, runtimeID, labelsForDBMock).Return(testErr).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			UIDServiceFn:   rtmtest.UnusedUUIDService(),
			WebhookService: rtmtest.UnusedWebhookService(),
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input: modelInput(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when parent label upserting failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctxWithSubaccount, subaccountID, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				svc := &automock.ScenariosService{}
				svc.On("AddDefaultScenarioIfEnabled", ctxWithSubaccount, subaccountID, &labelsForDBMockWithoutNormalization).Return().Once()
				return svc
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctxWithSubaccount, subaccountID, model.RuntimeLabelableObject, runtimeID, labelsForDBMockWithSubaccount).Return(nil).Once()
				repo.On("UpsertMultipleLabels", ctxWithGlobalaccountMatcher, tnt, model.RuntimeLabelableObject, runtimeID, parentScenarios).Return(testErr).Once()
				return repo
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, extSubaccountID).Return(&model.BusinessTenantMapping{ID: subaccountID, ExternalTenant: extSubaccountID, Parent: tnt}, nil).Once()
				tenantSvc.On("GetTenantByID", ctxWithSubaccountMatcher, subaccountID).Return(subaccount, nil).Once()
				return tenantSvc
			},
			UIDServiceFn: rtmtest.UnusedUUIDService(),
			WebhookService: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.Mock.On("Create", mock.Anything, runtimeID, webhookInput, model.RuntimeWebhookReference).Return("webhookID", nil)
				return webhookSvc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctxWithGlobalaccountMatcher, map[string]interface{}{}, runtimeID).Return([]interface{}{"test"}, nil)
				return svc
			},
			Input: modelInputWithSubaccountLabel(),
			MandatoryLabels: func() map[string]interface{} {
				return nilLabels
			},
			Context:     ctx,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RuntimeRepositoryFn()
			idSvc := testCase.UIDServiceFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			scenariosSvc := testCase.ScenariosServiceFn()
			engineSvc := testCase.EngineServiceFn()
			tenantSvc := testCase.TenantSvcFn()
			mandatoryLabels := testCase.MandatoryLabels()
			webhookSvc := testCase.WebhookService()
			svc := runtime.NewService(repo, nil, scenariosSvc, labelSvc, idSvc, engineSvc, tenantSvc, webhookSvc, protectedLabelPattern, immutableLabelPattern)

			// WHEN
			err := svc.CreateWithMandatoryLabels(testCase.Context, testCase.Input, runtimeID, mandatoryLabels)

			// then
			if err == nil {
				require.Nil(t, testCase.ExpectedErr)
			} else {
				require.NotNil(t, testCase.ExpectedErr)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, repo, idSvc, labelSvc, scenariosSvc, engineSvc, tenantSvc, webhookSvc)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		uuidSvc := &automock.UidService{}
		uuidSvc.On("Generate").Return(testUUID).Once()

		svc := runtime.NewService(nil, nil, nil, nil, uuidSvc, nil, nil, nil, protectedLabelPattern, immutableLabelPattern)
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
		"scenarios":               []interface{}{"SCENARIO"},
		runtime.IsNormalizedLabel: "true",
	}
	labels := map[string]interface{}{
		"label1": "val1",
	}
	protectedLabels := map[string]interface{}{
		"protected_defaultEventing": "true",
		"consumer_subaccount_ids":   []string{"subaccountID-1", "subaccountID-2"},
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
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelRepositoryFn    func() *automock.LabelRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		Input                model.RuntimeUpdateInput
		InputID              string
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, modelInput.Labels).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels, runtimeID).Return([]interface{}{}, nil)
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when updating with protected labels",
			RepositoryFn: func() *automock.RuntimeRepository {
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labels).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels, runtimeID).Return([]interface{}{}, nil)
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInputWithProtectedLabels,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when there are scenarios to set from assignments",
			RepositoryFn: func() *automock.RuntimeRepository {
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labelsDBMock).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels, runtimeID).Return([]interface{}{"SCENARIO"}, nil)
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when labels are nil",
			RepositoryFn: func() *automock.RuntimeRepository {
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labelsWithNormalization).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labelsWithNormalization, runtimeID).Return([]interface{}{}, nil)
				return svc
			},
			InputID: runtimeID,
			Input: model.RuntimeUpdateInput{
				Name: "bar",
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, tnt, inputRuntimeModel).Return(testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error if merge of scenarios and assignments failed",
			RepositoryFn: func() *automock.RuntimeRepository {
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels, runtimeID).Return(nil, testErr)
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when upserting labels failed",
			RepositoryFn: func() *automock.RuntimeRepository {
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, modelInput.Labels).Return(testErr).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels, runtimeID).Return([]interface{}{}, nil)
				return svc
			},
			InputID:            runtimeID,
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			engineSvc := testCase.EngineServiceFn()
			svc := runtime.NewService(repo, labelRepo, nil, labelSvc, nil, engineSvc, nil, nil, protectedLabelPattern, immutableLabelPattern)

			// WHEN
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
			engineSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "")
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

	desc := "Lorem ipsum"

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, runtimeModel.ID).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, runtimeModel.ID).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "")
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

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, nil, nil, "", "")

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
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "")
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

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, nil, nil, "", "")

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
			svc := runtime.NewService(rtmRepo, nil, nil, nil, nil, nil, nil, nil, "", "")

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
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "")
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
		fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime(t, "bar", "tenant-bar", "Bar", "Lorem Ipsum"),
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
			Name: "Returns error when pageSize is less than 1",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      0,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when pageSize is bigger than 200",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
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

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, nil, nil, "", "")

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
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "")
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
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
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
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil, nil, nil, nil, nil, "", "")

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
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "")
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

	secondProtectedModelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c13",
		Key:        "consumer_subaccount_ids",
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	labels := map[string]*model.Label{"protected_defaultEventing": protectedModelLabel, "consumer_subaccount_ids": secondProtectedModelLabel, "first": modelLabel, "second": modelLabel}
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
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
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
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern)

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
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, "", "")
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
	secondProtectedLabelKey := "consumer_subaccount_ids"

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

	secondModelProtectedLabelInput := model.LabelInput{
		Key:        secondProtectedLabelKey,
		Value:      []string{"value1", "value2"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	scenariosLabelValue := []interface{}{"SCENARIO"}
	modelScenariosLabelInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      scenariosLabelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	labelMapWithScenariosLabel := map[string]*model.Label{
		model.ScenariosKey: {
			ID:         "id",
			Tenant:     str.Ptr("tenant"),
			Key:        model.ScenariosKey,
			Value:      scenariosLabelValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelMap := map[string]*model.Label{
		labelKey: {
			ID:         "id",
			Key:        labelKey,
			Value:      []string{"val"},
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		LabelRepositoryFn    func() *automock.LabelRepository
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		InputRuntimeID       string
		InputLabel           *model.LabelInput
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelLabelInput).Return(nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label key is scenarios",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelScenariosLabelInput).Return(nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabel, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{model.ScenariosKey: scenariosLabelValue}, runtimeID).Return(scenariosLabelValue, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelScenariosLabelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if runtime exists failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime doesn't exists",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
		{
			Name: "Returns error when getting current labels for runtime failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label key is scenarios and merge scenarios and assignments failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabel, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{model.ScenariosKey: scenariosLabelValue}, runtimeID).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelScenariosLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns an error when trying to set protected label",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelProtectedLabelInput,
			ExpectedErrMessage: "could not set unmodifiable label with key protected_defaultEventing",
		},
		{
			Name: "Returns an error when trying to set consumer_subaccount_ids protected label",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &secondModelProtectedLabelInput,
			ExpectedErrMessage: "could not set unmodifiable label with key consumer_subaccount_ids",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			labelRepo := testCase.LabelRepositoryFn()
			engineSvc := testCase.EngineServiceFn()
			svc := runtime.NewService(repo, labelRepo, nil, labelSvc, nil, engineSvc, nil, nil, protectedLabelPattern, immutableLabelPattern)

			// WHEN
			err := svc.SetLabel(ctx, testCase.InputLabel)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			engineSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern)
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
	secondProtectedLabelKey := "consumer_subaccount_ids"
	labelValue := "val"
	labelKey2 := "key2"
	scenario := "SCENARIO"
	secondScenario := "SECOND_SCENARIO"
	scenariosLabelValue := []interface{}{scenario}
	scenariosLabelValueWithMultipleValues := []interface{}{scenario, secondScenario}

	labelMap := map[string]*model.Label{
		labelKey: {
			ID:         "id",
			Key:        labelKey,
			Value:      []string{"val"},
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelMapWithScenariosLabel := map[string]*model.Label{
		model.ScenariosKey: {
			ID:         "id",
			Tenant:     str.Ptr("tenant"),
			Key:        model.ScenariosKey,
			Value:      scenariosLabelValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelMapWithScenariosLabelWithMultipleValues := map[string]*model.Label{
		model.ScenariosKey: {
			ID:         "id",
			Tenant:     str.Ptr("tenant"),
			Key:        model.ScenariosKey,
			Value:      scenariosLabelValueWithMultipleValues,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
		labelKey: {
			ID:         "id",
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelSelectorValue := "selector"

	labelMapWithTwoSelectors := map[string]*model.Label{
		labelKey: {
			ID:         "id",
			Key:        labelKey,
			Value:      labelSelectorValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
		labelKey2: {
			ID:         "id",
			Key:        labelKey2,
			Value:      labelSelectorValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelRepositoryFn    func() *automock.LabelRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		InputRuntimeID       string
		InputKey             string
		ExpectedErrMessage   string
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
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label key is scenarios",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabelWithMultipleValues, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				modelLabelInput := &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []interface{}{scenario, secondScenario},
					ObjectID:   runtimeID,
					ObjectType: model.RuntimeLabelableObject,
				}
				svc.On("UpsertLabel", ctx, tnt, modelLabelInput).Return(nil).Once()
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{labelKey: labelValue}, runtimeID).Return(scenariosLabelValueWithMultipleValues, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label key is selector",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if runtime exists failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime does not exists",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
		{
			Name: "Returns error if listing current labels for runtime failed",
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
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label key is scenarios and merging scenarios and input assignments for old labels failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabel, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{}, runtimeID).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime scenario label delete failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime label delete failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when upserting scenarios label failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabelWithMultipleValues, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				modelLabelInput := &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []interface{}{scenario, secondScenario},
					ObjectID:   runtimeID,
					ObjectType: model.RuntimeLabelableObject,
				}
				svc.On("UpsertLabel", ctx, tnt, modelLabelInput).Return(testErr).Once()
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{labelKey: labelValue}, runtimeID).Return(scenariosLabelValueWithMultipleValues, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns an error when trying to delete protected label",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           protectedLabelKey,
			ExpectedErrMessage: "could not delete unmodifiable label with key protected_defaultEventing",
		},
		{
			Name: "Returns an error when trying to delete consumer_subaccount_ids protected label",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           secondProtectedLabelKey,
			ExpectedErrMessage: "could not delete unmodifiable label with key consumer_subaccount_ids",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			labelUpsertSvc := testCase.LabelUpsertServiceFn()
			engineSvc := testCase.EngineServiceFn()
			svc := runtime.NewService(repo, labelRepo, nil, labelUpsertSvc, nil, engineSvc, nil, nil, protectedLabelPattern, immutableLabelPattern)

			// WHEN
			err := svc.DeleteLabel(ctx, testCase.InputRuntimeID, testCase.InputKey)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			labelUpsertSvc.AssertExpectations(t)
			engineSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, nil, nil, protectedLabelPattern, immutableLabelPattern)
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
		&labelfilter.LabelFilter{Key: "test-key", Query: str.Ptr("test-filter")},
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
			labelUpsertService := &automock.LabelUpsertService{}
			scenariosService := &automock.ScenariosService{}
			scenarioAssignmentEngine := &automock.ScenarioAssignmentEngine{}
			uidSvc := &automock.UidService{}
			svc := runtime.NewService(repo, labelRepository, scenariosService, labelUpsertService, uidSvc, scenarioAssignmentEngine, nil, nil, protectedLabelPattern, immutableLabelPattern)

			// WHEN
			actualRuntime, err := svc.GetByFiltersGlobal(ctx, filters)
			// then
			if testCase.ExpectedErrMessage == "" {
				require.Equal(t, testRuntime, actualRuntime)
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepository.AssertExpectations(t)
			labelUpsertService.AssertExpectations(t)
			scenariosService.AssertExpectations(t)
			scenarioAssignmentEngine.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
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
		fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime(t, "bar", "tenant-bar", "Bar", "Lorem Ipsum"),
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
			labelUpsertService := &automock.LabelUpsertService{}
			scenariosService := &automock.ScenariosService{}
			scenarioAssignmentEngine := &automock.ScenarioAssignmentEngine{}
			uidSvc := &automock.UidService{}
			svc := runtime.NewService(repo, labelRepository, scenariosService, labelUpsertService, uidSvc, scenarioAssignmentEngine, nil, nil, protectedLabelPattern, immutableLabelPattern)

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

			repo.AssertExpectations(t)
			labelRepository.AssertExpectations(t)
			labelUpsertService.AssertExpectations(t)
			scenariosService.AssertExpectations(t)
			scenarioAssignmentEngine.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
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
		fixModelRuntime(t, "foo", tnt, "Foo", "Lorem Ipsum"),
		fixModelRuntime(t, "bar", tnt, "Bar", "Lorem Ipsum"),
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
			Name: "Fails when no tenant in the context",
			RepositoryFn: func() *automock.RuntimeRepository {
				return &automock.RuntimeRepository{}
			},
			Context:            context.TODO(),
			ExpectedErrMessage: "while loading tenant from context",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepository := &automock.LabelRepository{}
			labelUpsertService := &automock.LabelUpsertService{}
			scenariosService := &automock.ScenariosService{}
			scenarioAssignmentEngine := &automock.ScenarioAssignmentEngine{}
			uidSvc := &automock.UidService{}
			svc := runtime.NewService(repo, labelRepository, scenariosService, labelUpsertService, uidSvc, scenarioAssignmentEngine, nil, nil, ".*_defaultEventing$", immutableLabelPattern)

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

			repo.AssertExpectations(t)
			labelRepository.AssertExpectations(t)
			labelUpsertService.AssertExpectations(t)
			scenariosService.AssertExpectations(t)
			scenarioAssignmentEngine.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
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
