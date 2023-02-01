package formationtemplateconstraintreferences_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolver_AttachConstraintToFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                         string
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConstraintReferenceConverter func() *automock.ConstraintReferenceConverter
		ConstraintReferenceService   func() *automock.ConstraintReferenceService
		ExpectedOutput               *graphql.ConstraintReference
		ExpectedErrorMsg             string
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			ConstraintReferenceService: func() *automock.ConstraintReferenceService {
				svc := &automock.ConstraintReferenceService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), constraintReference).Return(nil)
				return svc
			},
			ConstraintReferenceConverter: func() *automock.ConstraintReferenceConverter {
				converter := &automock.ConstraintReferenceConverter{}
				converter.On("ToModel", gqlConstraintReference).Return(constraintReference, nil)
				return converter
			},
			ExpectedOutput:   gqlConstraintReference,
			ExpectedErrorMsg: "",
		},
		{
			Name: "Error wihle creating reference",
			TxFn: txGen.ThatDoesntExpectCommit,
			ConstraintReferenceService: func() *automock.ConstraintReferenceService {
				svc := &automock.ConstraintReferenceService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), constraintReference).Return(testErr)
				return svc
			},
			ConstraintReferenceConverter: func() *automock.ConstraintReferenceConverter {
				converter := &automock.ConstraintReferenceConverter{}
				converter.On("ToModel", gqlConstraintReference).Return(constraintReference, nil)
				return converter
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			ConstraintReferenceService: func() *automock.ConstraintReferenceService {
				svc := &automock.ConstraintReferenceService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), constraintReference).Return(nil)
				return svc
			},
			ConstraintReferenceConverter: func() *automock.ConstraintReferenceConverter {
				converter := &automock.ConstraintReferenceConverter{}
				converter.On("ToModel", gqlConstraintReference).Return(constraintReference, nil)
				return converter
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Returns error when failing on the beginning of a transaction",
			TxFn:             txGen.ThatFailsOnBegin,
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			constraintReferenceSvc := unusedConstraintReferenceService()
			if testCase.ConstraintReferenceService != nil {
				constraintReferenceSvc = testCase.ConstraintReferenceService()
			}
			constraintReferenceConverter := unusedConstraintReferenceConverter()
			if testCase.ConstraintReferenceConverter != nil {
				constraintReferenceConverter = testCase.ConstraintReferenceConverter()
			}
			resolver := formationtemplateconstraintreferences.NewResolver(transact, constraintReferenceConverter, constraintReferenceSvc)

			// WHEN
			result, err := resolver.AttachConstraintToFormationTemplate(ctx, constraintID, templateID)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, constraintReferenceSvc, constraintReferenceConverter)
		})
	}
}

func TestResolver_DetachConstraintFromFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConstraintReferenceService func() *automock.ConstraintReferenceService
		ExpectedOutput             *graphql.ConstraintReference
		ExpectedErrorMsg           string
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			ConstraintReferenceService: func() *automock.ConstraintReferenceService {
				svc := &automock.ConstraintReferenceService{}
				svc.On("Delete", txtest.CtxWithDBMatcher(), constraintID, templateID).Return(nil)
				return svc
			},
			ExpectedOutput:   gqlConstraintReference,
			ExpectedErrorMsg: "",
		},
		{
			Name: "Error when delete call in service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			ConstraintReferenceService: func() *automock.ConstraintReferenceService {
				svc := &automock.ConstraintReferenceService{}
				svc.On("Delete", txtest.CtxWithDBMatcher(), constraintID, templateID).Return(testErr)
				return svc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			ConstraintReferenceService: func() *automock.ConstraintReferenceService {
				svc := &automock.ConstraintReferenceService{}
				svc.On("Delete", txtest.CtxWithDBMatcher(), constraintID, templateID).Return(nil)
				return svc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Returns error when failing on the beginning of a transaction",
			TxFn:             txGen.ThatFailsOnBegin,
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			constraintReferenceSvc := unusedConstraintReferenceService()
			if testCase.ConstraintReferenceService != nil {
				constraintReferenceSvc = testCase.ConstraintReferenceService()
			}
			resolver := formationtemplateconstraintreferences.NewResolver(transact, nil, constraintReferenceSvc)

			// WHEN
			result, err := resolver.DetachConstraintFromFormationTemplate(ctx, constraintID, templateID)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, constraintReferenceSvc)
		})
	}
}
