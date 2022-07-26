package runtimectx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Exist(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")

	rtmCtxID := "id"

	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.RuntimeContextRepository
		InputRuntimeContextID string
		ExpectedValue         bool
		ExpectedError         error
	}{
		{
			Name: "RuntimeContext exists",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, rtmCtxID).Return(true, nil)
				return repo
			},
			InputRuntimeContextID: rtmCtxID,
			ExpectedValue:         true,
			ExpectedError:         nil,
		},
		{
			Name: "RuntimeContext not exits",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, rtmCtxID).Return(false, nil)
				return repo
			},
			InputRuntimeContextID: rtmCtxID,
			ExpectedValue:         false,
			ExpectedError:         nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, rtmCtxID).Return(false, testError)
				return repo
			},
			InputRuntimeContextID: rtmCtxID,
			ExpectedValue:         false,
			ExpectedError:         testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			rtmCtxRepo := testCase.RepositoryFn()
			svc := runtimectx.NewService(rtmCtxRepo, nil, nil, nil, nil, nil, nil)

			// WHEN
			value, err := svc.Exist(ctx, testCase.InputRuntimeContextID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, testCase.ExpectedValue, value)
			rtmCtxRepo.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtimectx.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.Exist(context.TODO(), "id")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	scenario := "scenario"
	id := "foo"
	runtimeID := "runtime_id"
	key := "key"
	val := "val"
	modelInput := model.RuntimeContextInput{
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	runtimeCtxModel := mock.MatchedBy(func(rtmCtx *model.RuntimeContext) bool {
		return rtmCtx.Key == modelInput.Key && rtmCtx.Value == modelInput.Value && rtmCtx.RuntimeID == modelInput.RuntimeID
	})

	tnt := "tenant"
	externalTnt := "external-tnt"
	parentTnt := "parent"
	modelTnt := &model.BusinessTenantMapping{ID: tnt, Parent: parentTnt}
	ctxWithoutTenant := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctxWithoutTenant, tnt, externalTnt)
	ctxWithParentTenant := tenant.SaveToContext(ctxWithTenant, parentTnt, "")
	formations := []string{"scenario"}

	testCases := []struct {
		Name                       string
		RuntimeContextRepositoryFn func() *automock.RuntimeContextRepository
		UIDServiceFn               func() *automock.UIDService
		TenantServiceFN            func() *automock.TenantService
		FormationServiceFn         func() *automock.FormationService
		RuntimeRepositoryFn        func() *automock.RuntimeRepository
		Input                      model.RuntimeContextInput
		Context                    context.Context
		ExpectedErrMessage         string
	}{
		{
			Name: "Success when runtime context's runtime have owner=false",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetScenariosFromMatchingASAs", ctxWithParentTenant, id, graphql.FormationObjectTypeRuntimeContext).Return(formations, nil).Once()
				formationSvc.On("AssignFormation", ctxWithParentTenant, parentTnt, id, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: scenario}).Return(nil, nil).Once()
				return formationSvc
			},
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("OwnerExists", ctxWithParentTenant, parentTnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(modelTnt, nil).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when runtime context's runtime have owner=true",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetScenariosFromMatchingASAs", ctxWithParentTenant, id, graphql.FormationObjectTypeRuntimeContext).Return(formations, nil).Once()
				formationSvc.On("AssignFormation", ctxWithParentTenant, parentTnt, id, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: scenario}).Return(nil, nil).Once()
				formationSvc.On("UnassignFormation", ctxWithParentTenant, parentTnt, runtimeID, graphql.FormationObjectTypeRuntime, model.Formation{Name: scenario}).Return(nil, nil).Once()
				return formationSvc
			},
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("OwnerExists", ctxWithParentTenant, parentTnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(modelTnt, nil).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when there aren't any scenarios from matching ASAs",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetScenariosFromMatchingASAs", ctxWithParentTenant, id, graphql.FormationObjectTypeRuntimeContext).Return([]string{}, nil).Once()
				return formationSvc
			},
			RuntimeRepositoryFn: unusedRuntimeRepo,
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(modelTnt, nil).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if runtime with owner=true exists fails",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetScenariosFromMatchingASAs", ctxWithParentTenant, id, graphql.FormationObjectTypeRuntimeContext).Return(formations, nil).Once()
				return formationSvc
			},
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("OwnerExists", ctxWithParentTenant, parentTnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(modelTnt, nil).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when unassign formation from runtime fails",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetScenariosFromMatchingASAs", ctxWithParentTenant, id, graphql.FormationObjectTypeRuntimeContext).Return(formations, nil).Once()
				formationSvc.On("AssignFormation", ctxWithParentTenant, parentTnt, id, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: scenario}).Return(nil, nil).Once()
				formationSvc.On("UnassignFormation", ctxWithParentTenant, parentTnt, runtimeID, graphql.FormationObjectTypeRuntime, model.Formation{Name: scenario}).Return(nil, testErr).Once()
				return formationSvc
			},
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("OwnerExists", ctxWithParentTenant, parentTnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(modelTnt, nil).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime context creation failed",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("").Once()
				return svc
			},
			FormationServiceFn:  unusedFormationService,
			TenantServiceFN:     unusedTenantService,
			RuntimeRepositoryFn: unusedRuntimeRepo,
			Input:               modelInput,
			Context:             ctxWithTenant,
			ExpectedErrMessage:  testErr.Error(),
		},
		{
			Name:                       "Returns error on loading tenant",
			RuntimeContextRepositoryFn: unusedRuntimeContextRepo,
			UIDServiceFn:               unusedUIDService,
			FormationServiceFn:         unusedFormationService,
			TenantServiceFN:            unusedTenantService,
			RuntimeRepositoryFn:        unusedRuntimeRepo,
			Input:                      model.RuntimeContextInput{},
			Context:                    ctxWithoutTenant,
			ExpectedErrMessage:         "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name: "Returns error when fetching tenant",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn:  unusedFormationService,
			RuntimeRepositoryFn: unusedRuntimeRepo,
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(nil, testErr).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "while getting tenant with id",
		},
		{
			Name: "Returns error when getting ASAs from parent",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetScenariosFromMatchingASAs", ctxWithParentTenant, id, graphql.FormationObjectTypeRuntimeContext).Return(nil, testErr).Once()
				return formationSvc
			},
			RuntimeRepositoryFn: unusedRuntimeRepo,
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(modelTnt, nil).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "while getting formations from automatic scenario assignments",
		},
		{
			Name: "Returns error while assigning runtime context to formation",
			RuntimeContextRepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Create", ctxWithTenant, tnt, runtimeCtxModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			FormationServiceFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetScenariosFromMatchingASAs", ctxWithParentTenant, id, graphql.FormationObjectTypeRuntimeContext).Return(formations, nil).Once()
				formationSvc.On("AssignFormation", ctxWithParentTenant, parentTnt, id, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: scenario}).Return(nil, testErr).Once()
				return formationSvc
			},
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("OwnerExists", ctxWithParentTenant, parentTnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			TenantServiceFN: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByID", ctxWithTenant, tnt).Return(modelTnt, nil).Once()
				return tenantSvc
			},
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "while assigning formation with name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RuntimeContextRepositoryFn()
			idSvc := testCase.UIDServiceFn()
			formationSvc := testCase.FormationServiceFn()
			tenantSvc := testCase.TenantServiceFN()
			runtimeRepo := testCase.RuntimeRepositoryFn()
			svc := runtimectx.NewService(repo, nil, runtimeRepo, nil, formationSvc, tenantSvc, idSvc)

			// WHEN
			result, err := svc.Create(testCase.Context, testCase.Input)

			// THEN

			if testCase.ExpectedErrMessage == "" {
				require.Nil(t, err)
				assert.IsType(t, "string", result)
			} else {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, formationSvc, tenantSvc, idSvc, runtimeRepo)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	modelInput := model.RuntimeContextInput{
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	inputRuntimeContextModel := mock.MatchedBy(func(rtmCtx *model.RuntimeContext) bool {
		return rtmCtx.Key == modelInput.Key && rtmCtx.Value == modelInput.Value && rtmCtx.RuntimeID == modelInput.RuntimeID
	})

	runtimeCtxModel := &model.RuntimeContext{
		ID:        id,
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctxWithoutTenant := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctxWithoutTenant, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeContextRepository
		Input              model.RuntimeContextInput
		InputID            string
		Context            context.Context
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenant, tnt, "foo").Return(runtimeCtxModel, nil).Once()
				repo.On("Update", ctxWithTenant, tnt, inputRuntimeContextModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime context update failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenant, tnt, "foo").Return(runtimeCtxModel, nil).Once()
				repo.On("Update", ctxWithTenant, tnt, inputRuntimeContextModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenant, tnt, "foo").Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			Input:              modelInput,
			Context:            ctxWithTenant,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when loading tenant from context failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				return repo
			},
			InputID:            id,
			Input:              model.RuntimeContextInput{},
			Context:            ctxWithoutTenant,
			ExpectedErrMessage: "while loading tenant from context: cannot read tenant from context",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := runtimectx.NewService(repo, nil, nil, nil, nil, nil, nil)

			// WHEN
			err := svc.Update(testCase.Context, testCase.InputID, testCase.Input)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	scenario := "scenario"
	runtimeCtxModel := &model.RuntimeContext{
		ID:        id,
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctxWithoutTenant := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctxWithoutTenant, tnt, externalTnt)

	formations := []string{"scenario"}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeContextRepository
		FormationServiceFN func() *automock.FormationService
		Input              model.RuntimeContextInput
		InputID            string
		Context            context.Context
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Delete", ctxWithTenant, tnt, runtimeCtxModel.ID).Return(nil).Once()
				return repo
			},
			FormationServiceFN: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetFormationsForObject", ctxWithTenant, tnt, model.RuntimeContextLabelableObject, id).Return(formations, nil).Once()
				formationSvc.On("UnassignFormation", ctxWithTenant, tnt, id, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: scenario}).Return(nil, nil).Once()
				return formationSvc
			},
			InputID:            id,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when loading tenant from context failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				return repo
			},
			FormationServiceFN: unusedFormationService,
			InputID:            id,
			Context:            ctxWithoutTenant,
			ExpectedErrMessage: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:         "Returns error when listing formations for runtime context",
			RepositoryFn: unusedRuntimeContextRepo,
			FormationServiceFN: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetFormationsForObject", ctxWithTenant, tnt, model.RuntimeContextLabelableObject, id).Return(nil, testErr).Once()
				return formationSvc
			},
			InputID:            id,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "while listing formations for runtime context with id",
		},
		{
			Name:         "Returns error while unassigning formation",
			RepositoryFn: unusedRuntimeContextRepo,
			FormationServiceFN: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetFormationsForObject", ctxWithTenant, tnt, model.RuntimeContextLabelableObject, id).Return(formations, nil).Once()
				formationSvc.On("UnassignFormation", ctxWithTenant, tnt, id, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: scenario}).Return(nil, testErr).Once()
				return formationSvc
			},
			InputID:            id,
			Context:            ctxWithTenant,
			ExpectedErrMessage: "while unassigning formation with name",
		},
		{
			Name: "Returns error when runtime context deletion failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Delete", ctxWithTenant, tnt, runtimeCtxModel.ID).Return(testErr).Once()
				return repo
			},
			FormationServiceFN: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetFormationsForObject", ctxWithTenant, tnt, model.RuntimeContextLabelableObject, id).Return(formations, nil).Once()
				formationSvc.On("UnassignFormation", ctxWithTenant, tnt, id, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: scenario}).Return(nil, nil).Once()
				return formationSvc
			},
			InputID:            id,
			Context:            ctxWithTenant,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			formationSvc := testCase.FormationServiceFN()
			svc := runtimectx.NewService(repo, nil, nil, nil, formationSvc, nil, nil)

			// WHEN
			err := svc.Delete(testCase.Context, testCase.InputID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, repo, formationSvc)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tnt := "tenant"
	externalTnt := "external-tnt"

	runtimeCtxModel := &model.RuntimeContext{
		ID:        id,
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                   string
		RepositoryFn           func() *automock.RuntimeContextRepository
		Input                  model.RuntimeContextInput
		InputID                string
		ExpectedRuntimeContext *model.RuntimeContext
		ExpectedErrMessage     string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(runtimeCtxModel, nil).Once()
				return repo
			},
			InputID:                id,
			ExpectedRuntimeContext: runtimeCtxModel,
			ExpectedErrMessage:     "",
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:                id,
			ExpectedRuntimeContext: runtimeCtxModel,
			ExpectedErrMessage:     testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtimectx.NewService(repo, nil, nil, nil, nil, nil, nil)

			// WHEN
			rtmCtx, err := svc.GetByID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntimeContext, rtmCtx)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtimectx.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetByID(context.TODO(), "id")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_GetForRuntime(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"
	tnt := "tenant"
	externalTnt := "external-tnt"

	runtimeCtxModel := &model.RuntimeContext{
		ID:        id,
		Key:       key,
		Value:     val,
		RuntimeID: runtimeID,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                   string
		RepositoryFn           func() *automock.RuntimeContextRepository
		Input                  model.RuntimeContextInput
		InputID                string
		ExpectedRuntimeContext *model.RuntimeContext
		ExpectedErrMessage     string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetForRuntime", ctx, tnt, id, runtimeID).Return(runtimeCtxModel, nil).Once()
				return repo
			},
			InputID:                id,
			ExpectedRuntimeContext: runtimeCtxModel,
			ExpectedErrMessage:     "",
		},
		{
			Name: "Returns error when runtime context retrieval failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetForRuntime", ctx, tnt, id, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			InputID:                id,
			ExpectedRuntimeContext: runtimeCtxModel,
			ExpectedErrMessage:     testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtimectx.NewService(repo, nil, nil, nil, nil, nil, nil)

			// WHEN
			rtmCtx, err := svc.GetForRuntime(ctx, testCase.InputID, runtimeID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntimeContext, rtmCtx)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtimectx.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetForRuntime(context.TODO(), "id", runtimeID)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_ListAllForRuntime(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	runtimeID := "runtime_id"

	id := "foo"
	key := "key"
	val := "value"

	runtimeContexts := []*model.RuntimeContext{
		{
			ID:        id,
			Key:       key,
			Value:     val,
			RuntimeID: runtimeID,
		},
	}

	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeContextRepository
		ExpectedResult     []*model.RuntimeContext
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListAllForRuntime", ctx, tnt, runtimeID).Return(runtimeContexts, nil).Once()
				return repo
			},
			ExpectedResult:     runtimeContexts,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime context listing failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListAllForRuntime", ctx, tnt, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := runtimectx.NewService(repo, nil, nil, nil, nil, nil, nil)

			// WHEN
			l, err := svc.ListAllForRuntime(ctx, runtimeID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedResult)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtimectx.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListAllForRuntime(context.TODO(), runtimeID)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_ListByFilter(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	runtimeID := "runtime_id"

	id := "foo"
	key := "key"
	val := "value"

	id2 := "bar"
	key2 := "key2"
	val2 := "value2"

	modelRuntimeContexts := []*model.RuntimeContext{
		{
			ID:        id,
			Key:       key,
			Value:     val,
			RuntimeID: runtimeID,
		},
		{
			ID:        id2,
			Key:       key2,
			Value:     val2,
			RuntimeID: runtimeID,
		},
	}
	runtimePage := &model.RuntimeContextPage{
		Data:       modelRuntimeContexts,
		TotalCount: len(modelRuntimeContexts),
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
		RepositoryFn       func() *automock.RuntimeContextRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      int
		InputCursor        string
		ExpectedResult     *model.RuntimeContextPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("List", ctx, runtimeID, tnt, filter, first, after).Return(runtimePage, nil).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     runtimePage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime context listing failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("List", ctx, runtimeID, tnt, filter, first, after).Return(nil, testErr).Once()
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
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
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
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
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

			svc := runtimectx.NewService(repo, nil, nil, nil, nil, nil, nil)

			// WHEN
			rtmCtx, err := svc.ListByFilter(ctx, runtimeID, testCase.InputLabelFilters, testCase.InputPageSize, testCase.InputCursor)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, rtmCtx)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtimectx.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByFilter(context.TODO(), "", nil, 1, "")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_ListByRuntimeIDs(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	runtimeID := "runtime_id"
	runtime2ID := "runtime2_id"

	runtimeIDs := []string{runtimeID, runtime2ID}

	id := "foo"
	key := "key"
	val := "value"

	id2 := "bar"
	key2 := "key2"
	val2 := "value2"

	modelRuntimeContexts := []*model.RuntimeContext{
		{
			ID:        id,
			Key:       key,
			Value:     val,
			RuntimeID: runtimeID,
		},
		{
			ID:        id2,
			Key:       key2,
			Value:     val2,
			RuntimeID: runtime2ID,
		},
	}
	runtimePage := &model.RuntimeContextPage{
		Data:       modelRuntimeContexts,
		TotalCount: len(modelRuntimeContexts),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"

	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeContextRepository
		InputPageSize      int
		InputCursor        string
		ExpectedResult     []*model.RuntimeContextPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByRuntimeIDs", ctx, tnt, runtimeIDs, first, after).Return([]*model.RuntimeContextPage{runtimePage}, nil).Once()
				return repo
			},
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     []*model.RuntimeContextPage{runtimePage},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime context listing failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByRuntimeIDs", ctx, tnt, runtimeIDs, first, after).Return(nil, testErr).Once()
				return repo
			},
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when pageSize is less than 1",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				return repo
			},
			InputPageSize:      0,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when pageSize is bigger than 200",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				return repo
			},
			InputPageSize:      201,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtimectx.NewService(repo, nil, nil, nil, nil, nil, nil)

			// WHEN
			rtmCtx, err := svc.ListByRuntimeIDs(ctx, runtimeIDs, testCase.InputPageSize, testCase.InputCursor)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, rtmCtx)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtimectx.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByRuntimeIDs(context.TODO(), runtimeIDs, 1, "")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_ListLabel(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeCtxID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}

	labels := map[string]*model.Label{"first": modelLabel, "second": modelLabel}
	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.RuntimeContextRepository
		LabelRepositoryFn     func() *automock.LabelRepository
		InputRuntimeContextID string
		InputLabel            *model.LabelInput
		ExpectedOutput        map[string]*model.Label
		ExpectedErrMessage    string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxID).Return(labels, nil).Once()
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedOutput:        labels,
			ExpectedErrMessage:    "",
		},
		{
			Name: "Returns error when labels receiving failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeContextLabelableObject, runtimeCtxID).Return(nil, testErr).Once()
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedOutput:        nil,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "Returns error when runtime context exists function failed",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedErrMessage:    testErr.Error(),
		},
		{
			Name: "Returns error when runtime context does not exists",
			RepositoryFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("Exists", ctx, tnt, runtimeCtxID).Return(false, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeContextID: runtimeCtxID,
			InputLabel:            label,
			ExpectedErrMessage:    fmt.Sprintf("runtime Context with ID %s doesn't exist", runtimeCtxID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtimectx.NewService(repo, labelRepo, nil, nil, nil, nil, nil)

			// WHEN
			l, err := svc.ListLabels(ctx, testCase.InputRuntimeContextID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedOutput)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// GIVEN
		svc := runtimectx.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListLabels(context.TODO(), "id")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func unusedRuntimeContextRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func unusedUIDService() *automock.UIDService {
	return &automock.UIDService{}
}

func unusedFormationService() *automock.FormationService {
	return &automock.FormationService{}
}

func unusedTenantService() *automock.TenantService {
	return &automock.TenantService{}
}

func unusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}
