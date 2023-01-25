package formation_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_EnsureScenarioAssigned(t *testing.T) {
	ctx := fixCtxWithTenant()

	testErr := errors.New("test err")

	rtmIDs := []string{"123", "456", "789"}
	rtmNames := []string{"first", "second", "third"}

	runtimes := []*model.Runtime{
		{
			ID:   rtmIDs[0],
			Name: rtmNames[0],
		},
		{
			ID:   rtmIDs[1],
			Name: rtmNames[1],
		},
		{
			ID:   rtmIDs[2],
			Name: rtmNames[2],
		},
	}

	ownedRuntimes := []*model.Runtime{runtimes[0], runtimes[1]}

	rtmContexts := []*model.RuntimeContext{
		{
			ID:        "1",
			RuntimeID: rtmIDs[0],
			Key:       "test",
			Value:     "test",
		},
		{
			ID:        "2",
			RuntimeID: rtmIDs[2],
			Key:       "test",
			Value:     "test",
		},
	}

	tnt := tenantID.String()
	testFormation := fixModel(testFormationName)

	testCases := []struct {
		Name                          string
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		ProcessFunc                   func() *automock.ProcessFunc
		InputASA                      model.AutomaticScenarioAssignment
		ExpectedErrMessage            string
	}{
		{
			Name: "Success",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, rtmContexts[1].ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name:                 "Returns error getting formation by name fails",
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(nil, testErr).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: unusedFormationTemplateRepo,
			InputASA:                      fixModel(testFormationName),
			ExpectedErrMessage:            testErr.Error(),
		},
		{
			Name:                 "Returns error getting formation template by ID fails",
			RuntimeRepoFN:        unusedRuntimeRepo,
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing owned runtimes fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: unusedRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime exists by id fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Times(1)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when assigning runtime to formation fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, testErr).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing all runtimes fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing runtime contexts for runtime fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when assigning runtime context to formation fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}

				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: testFormation.ScenarioName}).Return(nil, testErr).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
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
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepositoryFn != nil {
				formationRepo = testCase.FormationRepositoryFn()
			}
			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepositoryFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepositoryFn()
			}
			processFuncMock := unusedProcessFunc()
			if testCase.ProcessFunc != nil {
				processFuncMock = testCase.ProcessFunc()
			}

			svc := formation.NewASAEngine(nil, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, runtimeType, applicationType)

			// WHEN
			err := svc.EnsureScenarioAssigned(ctx, testCase.InputASA, processFuncMock.ProcessScenarioFunc)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, processFuncMock)
		})
	}
}

func TestService_RemoveAssignedScenario(t *testing.T) {
	ctx := fixCtxWithTenant()

	testErr := errors.New("test err")

	rtmIDs := []string{"123", "456", "789"}
	rtmNames := []string{"first", "second", "third"}

	runtimes := []*model.Runtime{
		{
			ID:   rtmIDs[0],
			Name: rtmNames[0],
		},
		{
			ID:   rtmIDs[1],
			Name: rtmNames[1],
		},
		{
			ID:   rtmIDs[2],
			Name: rtmNames[2],
		},
	}
	ownedRuntimes := []*model.Runtime{runtimes[0], runtimes[1]}

	rtmContexts := []*model.RuntimeContext{
		{
			ID:        "1",
			RuntimeID: rtmIDs[0],
			Key:       "test",
			Value:     "test",
		},
		{
			ID:        "2",
			RuntimeID: rtmIDs[2],
			Key:       "test",
			Value:     "test",
		},
	}

	tnt := tenantID.String()
	testFormation := fixModel(testFormationName)

	testCases := []struct {
		Name                          string
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		ProcessFunc                   func() *automock.ProcessFunc
		InputASA                      model.AutomaticScenarioAssignment
		ExpectedErrMessage            string
	}{
		{
			Name: "Success",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(nil, apperrors.NewNotFoundError(resource.RuntimeContext, rtmContexts[0].ID)).Once()
				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[2]).Return(rtmContexts[1], nil).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, rtmContexts[1].ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when processing scenarios for runtime context fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()

				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(rtmContexts[0], nil).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, rtmContexts[0].ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: testFormation.ScenarioName}).Return(nil, testErr).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when getting runtime context for runtime fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(runtimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				runtimeContextRepo.On("GetByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(nil, testErr).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing all runtimes fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				runtimeRepo.On("ListAllWithUnionSetCombination", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[1]).Return(true, nil).Once()

				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, nil).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when processing scenarios for runtime fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, nil).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ProcessFunc: func() *automock.ProcessFunc {
				f := &automock.ProcessFunc{}
				f.On("ProcessScenarioFunc", ctx, testFormation.Tenant, ownedRuntimes[0].ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: testFormation.ScenarioName}).Return(nil, testErr).Once()
				return f
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime context exists by parent runtime id fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(ownedRuntimes, nil).Once()
				return runtimeRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, rtmIDs[0]).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when listing owned runtimes fails",
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenantID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when getting formation template by id fails",
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(&modelFormation, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when getting formation by name fails",
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("GetByName", ctx, testFormationName, tnt).Return(nil, testErr).Once()
				return repo
			},
			InputASA:           fixModel(testFormationName),
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
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepositoryFn != nil {
				formationRepo = testCase.FormationRepositoryFn()
			}
			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepositoryFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepositoryFn()
			}
			processFuncMock := unusedProcessFunc()
			if testCase.ProcessFunc != nil {
				processFuncMock = testCase.ProcessFunc()
			}
			svc := formation.NewASAEngine(nil, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, runtimeType, applicationType)

			// WHEN
			err := svc.RemoveAssignedScenario(ctx, testCase.InputASA, processFuncMock.ProcessScenarioFunc)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			mock.AssertExpectationsForObjects(t, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, processFuncMock)
		})
	}
}

func TestService_GetScenariosFromMatchingASAs(t *testing.T) {
	ctx := fixCtxWithTenant()
	runtimeID := "runtimeID"
	runtimeID2 := "runtimeID2"

	testErr := errors.New(ErrMsg)
	notFoudErr := apperrors.NewNotFoundError(resource.Runtime, runtimeID2)

	testScenarios := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   ScenarioName,
			Tenant:         tenantID.String(),
			TargetTenantID: TargetTenantID,
		},
		{
			ScenarioName:   ScenarioName2,
			Tenant:         TenantID2,
			TargetTenantID: TargetTenantID2,
		},
	}

	formations := []*model.Formation{
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                ScenarioName,
		},
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                ScenarioName2,
		},
	}

	rtmCtx := &model.RuntimeContext{
		ID:        RuntimeContextID,
		Key:       "subscription",
		Value:     "subscriptionValue",
		RuntimeID: runtimeID,
	}

	rtmCtx2 := &model.RuntimeContext{
		ID:        RuntimeContextID,
		Key:       "subscription",
		Value:     "subscriptionValue",
		RuntimeID: runtimeID2,
	}

	testCases := []struct {
		Name                     string
		ScenarioAssignmentRepoFn func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFn            func() *automock.RuntimeRepository
		RuntimeContextRepoFn     func() *automock.RuntimeContextRepository
		FormationRepoFn          func() *automock.FormationRepository
		FormationTemplateRepoFn  func() *automock.FormationTemplateRepository
		ObjectID                 string
		ObjectType               graphql.FormationObjectType
		ExpectedErrorMessage     string
		ExpectedScenarios        []string
	}{
		{
			Name: "Success for runtime",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, RuntimeID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[1].TargetTenantID, RuntimeID, runtimeLblFilters).Return(false, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedErrorMessage: "",
			ExpectedScenarios:    []string{ScenarioName},
		},
		{
			Name: "Success for runtime context",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(rtmCtx, nil).Once()
				runtimeContextRepo.On("GetByID", ctx, testScenarios[1].TargetTenantID, RuntimeContextID).Return(rtmCtx2, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("GetByFiltersAndIDUsingUnion", ctx, testScenarios[0].TargetTenantID, rtmCtx.RuntimeID, runtimeLblFilters).Return(&model.Runtime{}, nil).Once()
				runtimeRepo.On("GetByFiltersAndIDUsingUnion", ctx, testScenarios[1].TargetTenantID, rtmCtx2.RuntimeID, runtimeLblFilters).Return(nil, notFoudErr).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeContextID,
			ObjectType:           graphql.FormationObjectTypeRuntimeContext,
			ExpectedErrorMessage: "",
			ExpectedScenarios:    []string{ScenarioName},
		},
		{
			Name: "Returns an error when getting runtime contexts",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(nil, testErr).Once()
				return runtimeContextRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeContextID,
			ObjectType:           graphql.FormationObjectTypeRuntimeContext,
			ExpectedErrorMessage: "while getting runtime contexts with ID",
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns an not found error when getting runtime contexts",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(nil, notFoudErr).Once()
				runtimeContextRepo.On("GetByID", ctx, testScenarios[1].TargetTenantID, RuntimeContextID).Return(nil, notFoudErr).Once()
				return runtimeContextRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeContextID,
			ObjectType:           graphql.FormationObjectTypeRuntimeContext,
			ExpectedErrorMessage: "",
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns an error when getting runtime",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("GetByID", ctx, testScenarios[0].TargetTenantID, RuntimeContextID).Return(rtmCtx, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("GetByFiltersAndIDUsingUnion", ctx, testScenarios[0].TargetTenantID, rtmCtx.RuntimeID, runtimeLblFilters).Return(nil, testErr).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeContextID,
			ObjectType:           graphql.FormationObjectTypeRuntimeContext,
			ExpectedErrorMessage: "while getting runtime with ID:",
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns an error when getting formations",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectID:             RuntimeContextID,
			ObjectType:           graphql.FormationObjectTypeRuntimeContext,
			ExpectedErrorMessage: "while getting formation by name",
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error for runtime when checking if the runtime has context fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, RuntimeID).Return(false, testErr).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(true, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedErrorMessage: "while cheking runtime context existence for runtime with ID",
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error for runtime when checking if runtime exists by filters and ID and has owner=true fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(false, testErr).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedErrorMessage: "while checking if asa matches runtime with ID rt-id: while checking if runtime with id \"rt-id\" have owner=true: some error",
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error for runtime when getting formation template runtime type fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(nil, testErr).Once()
				return formationTemplateRepo
			},
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedErrorMessage: "while checking if asa matches runtime with ID rt-id: while getting formation template by id \"bda5378d-caa1-4ee4-b8bf-f733e180fbf9\"",
			ExpectedScenarios:    nil,
		},
		{
			Name: "Returns error when listing ASAs fails",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(nil, testErr)
				return repo
			},
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedErrorMessage: "while listing Automatic Scenario Assignments in tenant",
			ExpectedScenarios:    nil,
		},
		{
			Name:                 "Returns error when can't find matching func",
			ObjectID:             "",
			ObjectType:           "test",
			ExpectedErrorMessage: "unexpected formation object type \"test\"",
			ExpectedScenarios:    nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := unusedASARepo()
			if testCase.ScenarioAssignmentRepoFn != nil {
				asaRepo = testCase.ScenarioAssignmentRepoFn()
			}
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFn != nil {
				runtimeRepo = testCase.RuntimeRepoFn()
			}
			runtimeContextRepo := unusedRuntimeContextRepo()
			if testCase.RuntimeContextRepoFn != nil {
				runtimeContextRepo = testCase.RuntimeContextRepoFn()
			}
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}
			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepoFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepoFn()
			}

			svc := formation.NewASAEngine(asaRepo, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, runtimeType, applicationType)

			// WHEN
			scenarios, err := svc.GetScenariosFromMatchingASAs(ctx, testCase.ObjectID, testCase.ObjectType)

			// THEN
			if testCase.ExpectedErrorMessage == "" {
				require.NoError(t, err)
				require.ElementsMatch(t, scenarios, testCase.ExpectedScenarios)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
				require.Nil(t, testCase.ExpectedScenarios)
			}

			mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo)
		})
	}
}

func TestService_IsFormationComingFromASA(t *testing.T) {
	ctx := fixCtxWithTenant()
	testScenarios := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   ScenarioName,
			Tenant:         tenantID.String(),
			TargetTenantID: TargetTenantID,
		},
		{
			ScenarioName:   ScenarioName2,
			Tenant:         TenantID2,
			TargetTenantID: TargetTenantID2,
		},
	}

	formations := []*model.Formation{
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                ScenarioName,
		},
		{
			ID:                  FormationID,
			TenantID:            tenantID.String(),
			FormationTemplateID: FormationTemplateID,
			Name:                ScenarioName2,
		},
	}

	testCases := []struct {
		Name                     string
		ScenarioAssignmentRepoFn func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFn            func() *automock.RuntimeRepository
		RuntimeContextRepoFn     func() *automock.RuntimeContextRepository
		FormationRepoFn          func() *automock.FormationRepository
		FormationTemplateRepoFn  func() *automock.FormationTemplateRepository
		FormationName            string
		ObjectID                 string
		ObjectType               graphql.FormationObjectType
		ExpectedErrorMessage     string
		ExpectedResult           bool
	}{
		{
			Name: "Success formation is coming from ASA",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, RuntimeID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[1].TargetTenantID, RuntimeID, runtimeLblFilters).Return(false, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			FormationName:        ScenarioName,
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedErrorMessage: "",
			ExpectedResult:       true,
		},
		{
			Name: "Success formation is not coming from ASA",
			ScenarioAssignmentRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				repo := &automock.AutomaticFormationAssignmentRepository{}
				repo.On("ListAll", ctx, tenantID.String()).Return(testScenarios, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				runtimeContextRepo := &automock.RuntimeContextRepository{}
				runtimeContextRepo.On("ExistsByRuntimeID", ctx, TargetTenantID, RuntimeID).Return(false, nil).Once()
				return runtimeContextRepo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[0].TargetTenantID, RuntimeID, runtimeLblFilters).Return(true, nil).Once()
				runtimeRepo.On("OwnerExistsByFiltersAndID", ctx, testScenarios[1].TargetTenantID, RuntimeID, runtimeLblFilters).Return(false, nil).Once()
				return runtimeRepo
			},
			FormationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, ScenarioName, testScenarios[0].Tenant).Return(formations[0], nil).Once()
				formationRepo.On("GetByName", ctx, ScenarioName2, testScenarios[1].Tenant).Return(formations[1], nil).Once()
				return formationRepo
			},
			FormationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctx, formations[0].FormationTemplateID).Return(&formationTemplate, nil).Once()
				formationTemplateRepo.On("Get", ctx, formations[1].FormationTemplateID).Return(&formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			FormationName:        ScenarioName2,
			ObjectID:             RuntimeID,
			ObjectType:           graphql.FormationObjectTypeRuntime,
			ExpectedErrorMessage: "",
			ExpectedResult:       false,
		},
		{
			Name:                 "Error when getting scenarios fails",
			ObjectID:             "",
			ObjectType:           "test",
			ExpectedErrorMessage: "unexpected formation object type \"test\"",
			ExpectedResult:       false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			asaRepo := unusedASARepo()
			if testCase.ScenarioAssignmentRepoFn != nil {
				asaRepo = testCase.ScenarioAssignmentRepoFn()
			}
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFn != nil {
				runtimeRepo = testCase.RuntimeRepoFn()
			}
			runtimeContextRepo := unusedRuntimeContextRepo()
			if testCase.RuntimeContextRepoFn != nil {
				runtimeContextRepo = testCase.RuntimeContextRepoFn()
			}
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepoFn != nil {
				formationRepo = testCase.FormationRepoFn()
			}
			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepoFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepoFn()
			}

			svc := formation.NewASAEngine(asaRepo, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, runtimeType, applicationType)

			// WHEN
			comesFromASA, err := svc.IsFormationComingFromASA(ctx, testCase.ObjectID, testCase.FormationName, testCase.ObjectType)

			// THEN
			if testCase.ExpectedErrorMessage == "" {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, comesFromASA)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
				require.False(t, comesFromASA)
			}

			mock.AssertExpectationsForObjects(t, asaRepo, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo)
		})
	}
}
