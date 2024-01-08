package formationassignment_test

import (
	"context"
	"testing"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_TargetEntity(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	emptyCtx = context.TODO()
	externalTnt = "externalTenant"
	ctxWithTenant = tenant.SaveToContext(emptyCtx, TestTenantID, externalTnt)

	appID := "appID"
	runtimeID := "runtimeID"
	runtimeContextID := "runtimeContextID"

	appModel := &model.Application{
		Name: "app1",
		BaseEntity: &model.BaseEntity{
			ID: appID,
		},
	}

	appGQL := &graphql.Application{
		Name: "app1",
		BaseEntity: &graphql.BaseEntity{
			ID: appID,
		},
	}

	appAssignment := &model.FormationAssignment{
		ID:          "ID1",
		FormationID: "ID",
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      appID,
		TargetType:  model.FormationAssignmentTypeApplication,
		State:       string(model.InitialAssignmentState),
	}

	rt := &model.Runtime{
		Name: "rt1",
		ID:   runtimeID,
	}

	rtGql := &graphql.Runtime{
		Name: "rt1",
		ID:   runtimeID,
	}

	rtAssignment := &model.FormationAssignment{
		ID:          "ID1",
		FormationID: "ID",
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      runtimeID,
		TargetType:  model.FormationAssignmentTypeRuntime,
		State:       string(model.InitialAssignmentState),
	}

	rtCtx := &model.RuntimeContext{
		ID: runtimeContextID,
	}

	rtCtxGql := &graphql.RuntimeContext{
		ID: runtimeContextID,
	}

	rtCtxAssignment := &model.FormationAssignment{
		ID:          "ID1",
		FormationID: "ID",
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      runtimeContextID,
		TargetType:  model.FormationAssignmentTypeRuntimeContext,
		State:       string(model.InitialAssignmentState),
	}

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AppRepo                 func() *automock.ApplicationRepo
		AppConverter            func() *automock.ApplicationConverter
		RuntimeRepo             func() *automock.RuntimeRepo
		RuntimeConverter        func() *automock.RuntimeConverter
		RuntimeContextRepo      func() *automock.RuntimeContextRepo
		RuntimeContextConverter func() *automock.RuntimeContextConverter
		ExpectedResult          []graphql.FormationParticipant
		ExpectedErr             []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			AppRepo: func() *automock.ApplicationRepo {
				repo := &automock.ApplicationRepo{}
				repo.On("ListAllByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{appID}).Return([]*model.Application{appModel}, nil).Once()
				return repo
			},
			AppConverter: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", appModel).Return(appGQL).Once()
				return conv
			},
			RuntimeRepo: func() *automock.RuntimeRepo {
				repo := &automock.RuntimeRepo{}
				repo.On("ListByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{runtimeID}).Return([]*model.Runtime{rt}, nil).Once()
				return repo
			},
			RuntimeConverter: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", rt).Return(rtGql).Once()
				return conv
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepo {
				repo := &automock.RuntimeContextRepo{}
				repo.On("ListByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{runtimeContextID}).Return([]*model.RuntimeContext{rtCtx}, nil).Once()
				return repo
			},
			RuntimeContextConverter: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("ToGraphQL", rtCtx).Return(rtCtxGql).Once()
				return conv
			},
			ExpectedResult: []graphql.FormationParticipant{appGQL, rtGql, rtCtxGql},
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedResult:  nil,
			ExpectedErr:     []error{testErr},
		},
		{
			Name:            "Error when listing Application participants",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			AppRepo: func() *automock.ApplicationRepo {
				repo := &automock.ApplicationRepo{}
				repo.On("ListAllByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{appID}).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Error when listing Runtime participants",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			AppRepo: func() *automock.ApplicationRepo {
				repo := &automock.ApplicationRepo{}
				repo.On("ListAllByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{appID}).Return([]*model.Application{appModel}, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepo {
				repo := &automock.RuntimeRepo{}
				repo.On("ListByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{runtimeID}).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Error when listing Runtime Context participants",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			AppRepo: func() *automock.ApplicationRepo {
				repo := &automock.ApplicationRepo{}
				repo.On("ListAllByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{appID}).Return([]*model.Application{appModel}, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepo {
				repo := &automock.RuntimeRepo{}
				repo.On("ListByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{runtimeID}).Return([]*model.Runtime{rt}, nil).Once()
				return repo
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepo {
				repo := &automock.RuntimeContextRepo{}
				repo.On("ListByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{runtimeContextID}).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			AppRepo: func() *automock.ApplicationRepo {
				repo := &automock.ApplicationRepo{}
				repo.On("ListAllByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{appID}).Return([]*model.Application{appModel}, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepo {
				repo := &automock.RuntimeRepo{}
				repo.On("ListByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{runtimeID}).Return([]*model.Runtime{rt}, nil).Once()
				return repo
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepo {
				repo := &automock.RuntimeContextRepo{}
				repo.On("ListByIDs", txtest.CtxWithDBMatcher(), TestTenantID, []string{runtimeContextID}).Return([]*model.RuntimeContext{rtCtx}, nil).Once()
				return repo
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			var appRepo *automock.ApplicationRepo
			if testCase.AppRepo != nil {
				appRepo = testCase.AppRepo()
			}
			var appConverter *automock.ApplicationConverter
			if testCase.AppConverter != nil {
				appConverter = testCase.AppConverter()
			}
			var rtRepo *automock.RuntimeRepo
			if testCase.RuntimeRepo != nil {
				rtRepo = testCase.RuntimeRepo()
			}
			var rtConverter *automock.RuntimeConverter
			if testCase.RuntimeConverter != nil {
				rtConverter = testCase.RuntimeConverter()
			}
			var rtContextRepo *automock.RuntimeContextRepo
			if testCase.RuntimeContextRepo != nil {
				rtContextRepo = testCase.RuntimeContextRepo()
			}
			var rtContextConverter *automock.RuntimeContextConverter
			if testCase.RuntimeContextConverter != nil {
				rtContextConverter = testCase.RuntimeContextConverter()
			}

			keys := []dataloader.ParamFormationParticipant{
				{
					ID:              appAssignment.ID,
					ParticipantID:   appID,
					ParticipantType: string(model.FormationAssignmentTypeApplication),
					Ctx:             ctxWithTenant,
				},
				{
					ID:              rtAssignment.ID,
					ParticipantID:   runtimeID,
					ParticipantType: string(model.FormationAssignmentTypeRuntime),
					Ctx:             ctxWithTenant,
				},
				{
					ID:              rtCtxAssignment.ID,
					ParticipantID:   runtimeContextID,
					ParticipantType: string(model.FormationAssignmentTypeRuntimeContext),
					Ctx:             ctxWithTenant,
				},
			}
			resolver := formationassignment.NewResolver(transact, appRepo, appConverter, rtRepo, rtConverter, rtContextRepo, rtContextConverter)

			// WHEN
			result, err := resolver.FormationParticipantDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			for i := range testCase.ExpectedErr {
				assert.Contains(t, err[i].Error(), testCase.ExpectedErr[i].Error())
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			if appRepo != nil {
				appRepo.AssertExpectations(t)
			}
			if appConverter != nil {
				appConverter.AssertExpectations(t)
			}
			if rtRepo != nil {
				rtRepo.AssertExpectations(t)
			}
			if rtConverter != nil {
				rtConverter.AssertExpectations(t)
			}
			if rtContextRepo != nil {
				rtContextRepo.AssertExpectations(t)
			}
			if rtContextConverter != nil {
				rtContextConverter.AssertExpectations(t)
			}
		})
	}

	t.Run("Returns error when there are no Assignments", func(t *testing.T) {
		resolver := formationassignment.NewResolver(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.FormationParticipantDataLoader([]dataloader.ParamFormationParticipant{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No Formation Assignments found").Error())
	})

	t.Run("Returns error when Participant ID is empty", func(t *testing.T) {
		params := dataloader.ParamFormationParticipant{ParticipantID: "", Ctx: context.TODO()}
		keys := []dataloader.ParamFormationParticipant{params}

		resolver := formationassignment.NewResolver(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.FormationParticipantDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("Cannot fetch Formation Participant. Participant ID is empty").Error())
	})
}
