package bundle_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_API(t *testing.T) {
	{
		// GIVEN
		id := "bar"
		bndlID := "1"
		var nilBundleID *string
		modelAPI := fixModelAPIDefinition(id, "name", "bar", "test")

		// TODO Revert when specs are fetched via subresolvers
		// modelSpec := &model.Spec{
		//	ID:         id,
		//	ObjectType: model.APISpecReference,
		//	ObjectID:   id,
		// }
		var modelSpec *model.Spec

		modelBundleRef := &model.BundleReference{
			BundleID:            &bndlID,
			ObjectType:          model.BundleAPIReference,
			ObjectID:            &id,
			APIDefaultTargetURL: str.Ptr(""),
		}
		gqlAPI := fixGQLAPIDefinition(id, bndlID, "name", "bar", "test")
		app := fixGQLBundle("foo", "foo", "foo")
		testErr := errors.New("Test error")
		txGen := txtest.NewTransactionContextGenerator(testErr)

		testCases := []struct {
			Name                     string
			TransactionerFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ServiceFn                func() *automock.APIService
			SpecServiceFn            func() *automock.SpecService
			BundleReferenceServiceFn func() *automock.BundleReferenceService
			ConverterFn              func() *automock.APIConverter
			InputID                  string
			Bundle                   *graphql.Bundle
			ExpectedAPI              *graphql.APIDefinition
			ExpectedErr              error
		}{
			{
				Name:            "Success",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleReferenceServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPI.ID, nilBundleID).Return(modelBundleRef, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI, modelSpec, modelBundleRef).Return(gqlAPI, nil).Once()
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: gqlAPI,
				ExpectedErr: nil,
			},
			{
				Name:            "Returns error when bundle retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				BundleReferenceServiceFn: func() *automock.BundleReferenceService {
					return &automock.BundleReferenceService{}
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns null when api for bundle not found",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()
					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				BundleReferenceServiceFn: func() *automock.BundleReferenceService {
					return &automock.BundleReferenceService{}
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: nil,
			},
			// TODO Revert when specs are fetched via subresolvers
			// {
			//	Name:            "Returns error when Spec retrieval failed",
			//	TransactionerFn: txGen.ThatDoesntExpectCommit,
			//	ServiceFn: func() *automock.APIService {
			//		svc := &automock.APIService{}
			//		svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()
			//
			//		return svc
			//	},
			//	SpecServiceFn: func() *automock.SpecService {
			//		svc := &automock.SpecService{}
			//		svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(nil, testErr).Once()
			//		return svc
			//	},
			//	BundleReferenceServiceFn: func() *automock.BundleReferenceService {
			//		return &automock.BundleReferenceService{}
			//	},
			//	ConverterFn: func() *automock.APIConverter {
			//		return &automock.APIConverter{}
			//	},
			//	InputID:     "foo",
			//	Bundle:      app,
			//	ExpectedAPI: nil,
			//	ExpectedErr: testErr,
			// },
			{
				Name:            "Returns error when BundleReference retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleReferenceServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPI.ID, nilBundleID).Return(nil, testErr).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when converting to GraphQL failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleReferenceServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPI.ID, nilBundleID).Return(modelBundleRef, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI, modelSpec, modelBundleRef).Return(nil, testErr).Once()
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when commit begin error",
				TransactionerFn: txGen.ThatFailsOnBegin,
				ServiceFn: func() *automock.APIService {
					return &automock.APIService{}
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				BundleReferenceServiceFn: func() *automock.BundleReferenceService {
					return &automock.BundleReferenceService{}
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when commit failed",
				TransactionerFn: txGen.ThatFailsOnCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleReferenceServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &modelAPI.ID, nilBundleID).Return(modelBundleRef, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI, modelSpec, modelBundleRef).Return(gqlAPI, nil).Once()
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				persist, transact := testCase.TransactionerFn()
				svc := testCase.ServiceFn()
				converter := testCase.ConverterFn()
				specSvc := testCase.SpecServiceFn()
				bndlRefSvc := testCase.BundleReferenceServiceFn()

				resolver := bundle.NewResolver(transact, nil, nil, bndlRefSvc, svc, nil, nil, nil, nil, converter, nil, nil, specSvc, nil)

				// WHEN
				result, err := resolver.APIDefinition(context.TODO(), testCase.Bundle, testCase.InputID)

				// then
				assert.Equal(t, testCase.ExpectedAPI, result)
				if testCase.ExpectedErr != nil {
					require.Error(t, err)
					assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				} else {
					require.Nil(t, err)
				}

				svc.AssertExpectations(t)
				persist.AssertExpectations(t)
				transact.AssertExpectations(t)
				converter.AssertExpectations(t)
				specSvc.AssertExpectations(t)
				bndlRefSvc.AssertExpectations(t)
			})
		}
	}
}

func TestResolver_APIs(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")
	group := "group"
	desc := "desc"
	name := "test name"

	firstBundleID := "bundleID"
	secondBundleID := "bundleID2"
	bundleIDs := []string{firstBundleID, secondBundleID}
	firstAPIID := "apiID"
	secondAPIID := "apiID2"
	// TODO Revert when specs are fetched via subresolvers
	// apiIDs := []string{firstAPIID, secondAPIID}
	// firstSpecID := "specID"
	// secondSpecID := "specID2"

	// model APIDefs
	apiDefFirstBundle := fixModelAPIDefinition(firstAPIID, "Foo", "Lorem Ipsum", group)
	apiDefSecondBundle := fixModelAPIDefinition(secondAPIID, "Bar", "Lorem Ipsum", group)

	apiDefsFirstBundle := []*model.APIDefinition{apiDefFirstBundle}
	apiDefsSecondBundle := []*model.APIDefinition{apiDefSecondBundle}

	apiDefPageFirstBundle := fixAPIDefinitionPage(apiDefsFirstBundle)
	apiDefPageSecondBundle := fixAPIDefinitionPage(apiDefsSecondBundle)
	apiDefPages := []*model.APIDefinitionPage{apiDefPageFirstBundle, apiDefPageSecondBundle}

	// GQL APIDefs
	gqlAPIDefFirstBundle := fixGQLAPIDefinition(firstAPIID, firstBundleID, name, desc, group)
	gqlAPIDefSecondBundle := fixGQLAPIDefinition(secondAPIID, secondBundleID, name, desc, group)

	gqlAPIDefsFirstBundle := []*graphql.APIDefinition{gqlAPIDefFirstBundle}
	gqlAPIDefsSecondBundle := []*graphql.APIDefinition{gqlAPIDefSecondBundle}

	gqlAPIDefPageFirstBundle := fixGQLAPIDefinitionPage(gqlAPIDefsFirstBundle)
	gqlAPIDefPageSecondBundle := fixGQLAPIDefinitionPage(gqlAPIDefsSecondBundle)
	gqlAPIDefPages := []*graphql.APIDefinitionPage{gqlAPIDefPageFirstBundle, gqlAPIDefPageSecondBundle}

	// API BundleReferences
	numberOfAPIsInFirstBundle := 1
	numberOfAPIsInSecondBundle := 1
	apiDefFirstBundleReference := fixModelAPIBundleReference(firstBundleID, firstAPIID)
	apiDefSecondBundleReference := fixModelAPIBundleReference(secondBundleID, secondAPIID)
	bundleRefsFirstAPI := []*model.BundleReference{apiDefFirstBundleReference}
	bundleRefsSecondAPI := []*model.BundleReference{apiDefSecondBundleReference}
	bundleRefs := []*model.BundleReference{apiDefFirstBundleReference, apiDefSecondBundleReference}
	totalCounts := map[string]int{firstBundleID: numberOfAPIsInFirstBundle, secondBundleID: numberOfAPIsInSecondBundle}

	// API Specs
	// TODO Revert when specs are fetched via subresolvers
	// apiDefFirstSpec := &model.Spec{ID: firstSpecID, ObjectType: model.APISpecReference, ObjectID: firstAPIID}
	// apiDefSecondSpec := &model.Spec{ID: secondSpecID, ObjectType: model.APISpecReference, ObjectID: secondAPIID}
	// specs := []*model.Spec{nil}
	specsFirstAPI := []*model.Spec{nil}
	specsSecondAPI := []*model.Spec{nil}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.APIService
		ConverterFn       func() *automock.APIConverter
		SpecServiceFn     func() *automock.SpecService
		BundleReferenceFn func() *automock.BundleReferenceService
		ExpectedResult    []*graphql.APIDefinitionPage
		ExpectedErr       []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(apiDefPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.APISpecReference, apiIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleAPIReference, bundleIDs, first, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", apiDefsFirstBundle, specsFirstAPI, bundleRefsFirstAPI).Return(gqlAPIDefsFirstBundle, nil).Once()
				conv.On("MultipleToGraphQL", apiDefsSecondBundle, specsSecondAPI, bundleRefsSecondAPI).Return(gqlAPIDefsSecondBundle, nil).Once()
				return conv
			},
			ExpectedResult: gqlAPIDefPages,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when APIs listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(nil, testErr).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		// TODO Revert when specs are fetched via subresolvers
		// {
		//	Name:            "Returns error when Specs retrieval failed",
		//	TransactionerFn: txGen.ThatDoesntExpectCommit,
		//	ServiceFn: func() *automock.APIService {
		//		svc := &automock.APIService{}
		//		svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(apiDefPages, nil).Once()
		//		return svc
		//	},
		//	SpecServiceFn: func() *automock.SpecService {
		//		svc := &automock.SpecService{}
		//		svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.APISpecReference, apiIDs).Return(nil, testErr).Once()
		//		return svc
		//	},
		//	BundleReferenceFn: func() *automock.BundleReferenceService {
		//		return &automock.BundleReferenceService{}
		//	},
		//	ConverterFn: func() *automock.APIConverter {
		//		return &automock.APIConverter{}
		//	},
		//	ExpectedResult: nil,
		//	ExpectedErr:    []error{testErr},
		// },
		{
			Name:            "Returns error when BundleReferences retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(apiDefPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.APISpecReference, apiIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleAPIReference, bundleIDs, first, after).Return(nil, nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when there is no BundleReference for API",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(apiDefPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.APISpecReference, apiIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				invalidBundleRefs := []*model.BundleReference{apiDefSecondBundleReference}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleAPIReference, bundleIDs, first, after).Return(invalidBundleRefs, totalCounts, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{errors.New("could not find BundleReference for API with id")},
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(apiDefPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.APISpecReference, apiIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleAPIReference, bundleIDs, first, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", apiDefsFirstBundle, specsFirstAPI, bundleRefsFirstAPI).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(apiDefPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.APISpecReference, apiIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleAPIReference, bundleIDs, first, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", apiDefsFirstBundle, specsFirstAPI, bundleRefsFirstAPI).Return(gqlAPIDefsFirstBundle, nil).Once()
				conv.On("MultipleToGraphQL", apiDefsSecondBundle, specsSecondAPI, bundleRefsSecondAPI).Return(gqlAPIDefsSecondBundle, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			specService := testCase.SpecServiceFn()
			bundleRefService := testCase.BundleReferenceFn()

			firstBundleParams := dataloader.ParamAPIDef{ID: firstBundleID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			secondBundleParams := dataloader.ParamAPIDef{ID: secondBundleID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			keys := []dataloader.ParamAPIDef{firstBundleParams, secondBundleParams}
			resolver := bundle.NewResolver(transact, nil, nil, bundleRefService, svc, nil, nil, nil, nil, converter, nil, nil, specService, nil)
			// WHEN
			result, err := resolver.APIDefinitionsDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err[0])
				assert.Contains(t, err[0].Error(), testCase.ExpectedErr[0].Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			specService.AssertExpectations(t)
			bundleRefService.AssertExpectations(t)
		})
	}

	t.Run("Returns error when there are no Bundles", func(t *testing.T) {
		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.APIDefinitionsDataLoader([]dataloader.ParamAPIDef{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No Bundles found").Error())
	})

	t.Run("Returns error when start cursor is nil", func(t *testing.T) {
		params := dataloader.ParamAPIDef{ID: firstBundleID, Ctx: context.TODO(), First: nil, After: &gqlAfter}
		keys := []dataloader.ParamAPIDef{params}

		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.APIDefinitionsDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInvalidDataError("missing required parameter 'first'").Error())
	})
}

func TestResolver_Event(t *testing.T) {
	{
		// GIVEN
		id := "bar"
		bndlID := "1"
		var nilBundleID *string
		modelEvent := fixModelEventAPIDefinition(id, "name", "bar", "test")
		// TODO Revert when specs are fetched via subresolvers
		// modelSpec := &model.Spec{
		//	ID:         id,
		//	ObjectType: model.EventSpecReference,
		//	ObjectID:   id,
		// }
		var modelSpec *model.Spec

		modelBundleRef := &model.BundleReference{
			BundleID:   &bndlID,
			ObjectType: model.BundleEventReference,
			ObjectID:   &id,
		}
		gqlEvent := fixGQLEventDefinition(id, bndlID, "name", "bar", "test")
		app := fixGQLBundle("foo", "foo", "foo")
		testErr := errors.New("Test error")
		txGen := txtest.NewTransactionContextGenerator(testErr)

		testCases := []struct {
			Name               string
			TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ServiceFn          func() *automock.EventService
			SpecServiceFn      func() *automock.SpecService
			BundleRefServiceFn func() *automock.BundleReferenceService
			ConverterFn        func() *automock.EventConverter
			InputID            string
			Bundle             *graphql.Bundle
			ExpectedEvent      *graphql.EventDefinition
			ExpectedErr        error
		}{
			{
				Name:            "Success",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleRefServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleEventReference, &modelEvent.ID, nilBundleID).Return(modelBundleRef, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					conv := &automock.EventConverter{}
					conv.On("ToGraphQL", modelEvent, modelSpec, modelBundleRef).Return(gqlEvent, nil).Once()
					return conv
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: gqlEvent,
				ExpectedErr:   nil,
			},
			{
				Name:            "Returns error when bundle retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				BundleRefServiceFn: func() *automock.BundleReferenceService {
					return &automock.BundleReferenceService{}
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns null when api for bundle not found",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()
					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				BundleRefServiceFn: func() *automock.BundleReferenceService {
					return &automock.BundleReferenceService{}
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   nil,
			},
			// TODO Revert when specs are fetched via subresolvers
			// {
			//	Name:            "Returns error when Spec retrieval failed",
			//	TransactionerFn: txGen.ThatDoesntExpectCommit,
			//	ServiceFn: func() *automock.EventService {
			//		svc := &automock.EventService{}
			//		svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()
			//
			//		return svc
			//	},
			//	SpecServiceFn: func() *automock.SpecService {
			//		svc := &automock.SpecService{}
			//		svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(nil, testErr).Once()
			//		return svc
			//	},
			//	BundleRefServiceFn: func() *automock.BundleReferenceService {
			//		return &automock.BundleReferenceService{}
			//	},
			//	ConverterFn: func() *automock.EventConverter {
			//		return &automock.EventConverter{}
			//	},
			//	InputID:       "foo",
			//	Bundle:        app,
			//	ExpectedEvent: nil,
			//	ExpectedErr:   testErr,
			// },
			{
				Name:            "Returns error when BundleReference retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleRefServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleEventReference, &modelEvent.ID, nilBundleID).Return(nil, testErr).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns error when converting to GraphQL failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleRefServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleEventReference, &modelEvent.ID, nilBundleID).Return(modelBundleRef, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					conv := &automock.EventConverter{}
					conv.On("ToGraphQL", modelEvent, modelSpec, modelBundleRef).Return(nil, testErr).Once()
					return conv
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns error when commit begin error",
				TransactionerFn: txGen.ThatFailsOnBegin,
				ServiceFn: func() *automock.EventService {
					return &automock.EventService{}
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				BundleRefServiceFn: func() *automock.BundleReferenceService {
					return &automock.BundleReferenceService{}
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns error when commit failed",
				TransactionerFn: txGen.ThatFailsOnCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					// TODO Revert when specs are fetched via subresolvers
					// svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(modelSpec, nil).Once()
					return svc
				},
				BundleRefServiceFn: func() *automock.BundleReferenceService {
					svc := &automock.BundleReferenceService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), model.BundleEventReference, &modelEvent.ID, nilBundleID).Return(modelBundleRef, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					conv := &automock.EventConverter{}
					conv.On("ToGraphQL", modelEvent, modelSpec, modelBundleRef).Return(gqlEvent, nil).Once()
					return conv
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				persist, transact := testCase.TransactionerFn()
				svc := testCase.ServiceFn()
				converter := testCase.ConverterFn()
				specSvc := testCase.SpecServiceFn()
				bndlRefService := testCase.BundleRefServiceFn()

				resolver := bundle.NewResolver(transact, nil, nil, bndlRefService, nil, svc, nil, nil, nil, nil, converter, nil, specSvc, nil)

				// WHEN
				result, err := resolver.EventDefinition(context.TODO(), testCase.Bundle, testCase.InputID)

				// then
				assert.Equal(t, testCase.ExpectedEvent, result)
				if testCase.ExpectedErr != nil {
					require.Error(t, err)
					assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				} else {
					require.Nil(t, err)
				}

				svc.AssertExpectations(t)
				persist.AssertExpectations(t)
				transact.AssertExpectations(t)
				converter.AssertExpectations(t)
				specSvc.AssertExpectations(t)
				bndlRefService.AssertExpectations(t)
			})
		}
	}
}

func TestResolver_Events(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")
	group := "group"
	desc := "desc"
	name := "test name"

	firstBundleID := "bundleID"
	secondBundleID := "bundleID2"
	bundleIDs := []string{firstBundleID, secondBundleID}
	firstEventID := "eventID"
	secondEventID := "eventID2"
	// TODO Revert when specs are fetched via subresolvers
	// eventIDs := []string{firstEventID, secondEventID}
	// firstSpecID := "specID"
	// secondSpecID := "specID2"

	// model Events
	eventFirstBundle := fixModelEventAPIDefinition(firstEventID, "Foo", "Lorem Ipsum", group)
	eventSecondBundle := fixModelEventAPIDefinition(secondEventID, "Bar", "Lorem Ipsum", group)

	eventsFirstBundle := []*model.EventDefinition{eventFirstBundle}
	eventsSecondBundle := []*model.EventDefinition{eventSecondBundle}

	eventPageFirstBundle := fixEventAPIDefinitionPage(eventsFirstBundle)
	eventPageSecondBundle := fixEventAPIDefinitionPage(eventsSecondBundle)
	eventPages := []*model.EventDefinitionPage{eventPageFirstBundle, eventPageSecondBundle}

	// GQL Events
	gqlEventFirstBundle := fixGQLEventDefinition(firstEventID, firstBundleID, name, desc, group)
	gqlEventSecondBundle := fixGQLEventDefinition(secondEventID, secondBundleID, name, desc, group)

	gqlEventsFirstBundle := []*graphql.EventDefinition{gqlEventFirstBundle}
	gqlEventsSecondBundle := []*graphql.EventDefinition{gqlEventSecondBundle}

	gqlEventPageFirstBundle := fixGQLEventDefinitionPage(gqlEventsFirstBundle)
	gqlEventPageSecondBundle := fixGQLEventDefinitionPage(gqlEventsSecondBundle)
	gqlEventPages := []*graphql.EventDefinitionPage{gqlEventPageFirstBundle, gqlEventPageSecondBundle}

	// Event BundleReferences
	numberOfEventsInFirstBundle := 1
	numberOfEventsInSecondBundle := 1
	eventFirstBundleReference := fixModelEventBundleReference(firstBundleID, firstEventID)
	eventSecondBundleReference := fixModelEventBundleReference(secondBundleID, secondEventID)
	bundleRefsFirstEvent := []*model.BundleReference{eventFirstBundleReference}
	bundleRefsSecondEvent := []*model.BundleReference{eventSecondBundleReference}
	bundleRefs := []*model.BundleReference{eventFirstBundleReference, eventSecondBundleReference}
	totalCounts := map[string]int{firstBundleID: numberOfEventsInFirstBundle, secondBundleID: numberOfEventsInSecondBundle}

	// Event Specs
	// TODO Revert when specs are fetched via subresolvers
	// eventFirstSpec := &model.Spec{ID: firstSpecID, ObjectType: model.EventSpecReference, ObjectID: firstEventID}
	// eventSecondSpec := &model.Spec{ID: secondSpecID, ObjectType: model.EventSpecReference, ObjectID: secondEventID}
	// specsFirstEvent := []*model.Spec{eventFirstSpec}
	// specsSecondEvent := []*model.Spec{eventSecondSpec}
	// specs := []*model.Spec{eventFirstSpec, eventSecondSpec}
	specsFirstEvent := []*model.Spec{nil}
	specsSecondEvent := []*model.Spec{nil}
	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.EventService
		ConverterFn       func() *automock.EventConverter
		SpecServiceFn     func() *automock.SpecService
		BundleReferenceFn func() *automock.BundleReferenceService
		ExpectedResult    []*graphql.EventDefinitionPage
		ExpectedErr       []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(eventPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleEventReference, bundleIDs, first, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleToGraphQL", eventsFirstBundle, specsFirstEvent, bundleRefsFirstEvent).Return(gqlEventsFirstBundle, nil).Once()
				conv.On("MultipleToGraphQL", eventsSecondBundle, specsSecondEvent, bundleRefsSecondEvent).Return(gqlEventsSecondBundle, nil).Once()
				return conv
			},
			ExpectedResult: gqlEventPages,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventService {
				return &automock.EventService{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.EventConverter {
				return &automock.EventConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when Events listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(nil, testErr).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ConverterFn: func() *automock.EventConverter {
				return &automock.EventConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		// TODO Revert when specs are fetched via subresolvers
		// {
		//	Name:            "Returns error when Specs retrieval failed",
		//	TransactionerFn: txGen.ThatDoesntExpectCommit,
		//	ServiceFn: func() *automock.EventService {
		//		svc := &automock.EventService{}
		//		svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(eventPages, nil).Once()
		//		return svc
		//	},
		//	SpecServiceFn: func() *automock.SpecService {
		//		svc := &automock.SpecService{}
		//		svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventIDs).Return(nil, testErr).Once()
		//		return svc
		//	},
		//	BundleReferenceFn: func() *automock.BundleReferenceService {
		//		return &automock.BundleReferenceService{}
		//	},
		//	ConverterFn: func() *automock.EventConverter {
		//		return &automock.EventConverter{}
		//	},
		//	ExpectedResult: nil,
		//	ExpectedErr:    []error{testErr},
		// },
		{
			Name:            "Returns error when BundleReferences retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(eventPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleEventReference, bundleIDs, first, after).Return(nil, nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				return &automock.EventConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(eventPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleEventReference, bundleIDs, first, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleToGraphQL", eventsFirstBundle, specsFirstEvent, bundleRefsFirstEvent).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), bundleIDs, first, after).Return(eventPages, nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				// TODO Revert when specs are fetched via subresolvers
				// svc.On("ListByReferenceObjectIDs", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventIDs).Return(specs, nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", txtest.CtxWithDBMatcher(), model.BundleEventReference, bundleIDs, first, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleToGraphQL", eventsFirstBundle, specsFirstEvent, bundleRefsFirstEvent).Return(gqlEventsFirstBundle, nil).Once()
				conv.On("MultipleToGraphQL", eventsSecondBundle, specsSecondEvent, bundleRefsSecondEvent).Return(gqlEventsSecondBundle, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			specService := testCase.SpecServiceFn()
			bundleRefService := testCase.BundleReferenceFn()

			firstBundleParams := dataloader.ParamEventDef{ID: firstBundleID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			secondBundleParams := dataloader.ParamEventDef{ID: secondBundleID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			keys := []dataloader.ParamEventDef{firstBundleParams, secondBundleParams}
			resolver := bundle.NewResolver(transact, nil, nil, bundleRefService, nil, svc, nil, nil, nil, nil, converter, nil, specService, nil)
			// WHEN
			result, err := resolver.EventDefinitionsDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err[0])
				assert.Contains(t, err[0].Error(), testCase.ExpectedErr[0].Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			specService.AssertExpectations(t)
			bundleRefService.AssertExpectations(t)
		})
	}

	t.Run("Returns error when there are no Bundles", func(t *testing.T) {
		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.EventDefinitionsDataLoader([]dataloader.ParamEventDef{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No Bundles found").Error())
	})

	t.Run("Returns error when start cursor is nil", func(t *testing.T) {
		params := dataloader.ParamEventDef{ID: firstBundleID, Ctx: context.TODO(), First: nil, After: &gqlAfter}
		keys := []dataloader.ParamEventDef{params}

		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.EventDefinitionsDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInvalidDataError("missing required parameter 'first'").Error())
	})
}

func TestResolver_Document(t *testing.T) {
	// GIVEN
	id := "bar"

	modelDoc := fixModelDocument("foo", id)
	gqlDoc := fixGQLDocument(id)
	bndl := fixGQLBundle("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.DocumentConverter
		InputID         string
		Bundle          *graphql.Bundle
		ExpectedDoc     *graphql.Document
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelDoc, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("ToGraphQL", modelDoc).Return(gqlDoc).Once()
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: gqlDoc,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns null when document for bundle not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelDoc, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := bundle.NewResolver(transact, nil, nil, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil)

			// WHEN
			result, err := resolver.Document(context.TODO(), testCase.Bundle, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedDoc, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Documents(t *testing.T) {
	// GIVEN
	contextParam := txtest.CtxWithDBMatcher()

	firstBundleID := "bundleID"
	secondBundleID := "bundleID2"
	bundleIDs := []string{firstBundleID, secondBundleID}
	firstDocID := "docID"
	secondDocID := "docID2"

	// model Docs
	docFirstBundle := fixModelDocument(firstBundleID, firstDocID)
	docSecondBundle := fixModelDocument(secondBundleID, secondDocID)

	docsFirstBundle := []*model.Document{docFirstBundle}
	docsSecondBundle := []*model.Document{docSecondBundle}

	docPageFirstBundle := fixModelDocumentPage(docsFirstBundle)
	docPageSecondBundle := fixModelDocumentPage(docsSecondBundle)
	docPages := []*model.DocumentPage{docPageFirstBundle, docPageSecondBundle}

	// GQL Docs
	gqlDocFirstBundle := fixGQLDocument(firstDocID)
	gqlDocSecondBundle := fixGQLDocument(secondDocID)

	gqlDocsFirstBundle := []*graphql.Document{gqlDocFirstBundle}
	gqlDocsSecondBundle := []*graphql.Document{gqlDocSecondBundle}

	gqlDocPageFirstBundle := fixGQLDocumentPage(gqlDocsFirstBundle)
	gqlDocPageSecondBundle := fixGQLDocumentPage(gqlDocsSecondBundle)
	gqlDocPages := []*graphql.DocumentPage{gqlDocPageFirstBundle, gqlDocPageSecondBundle}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.DocumentConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedResult  []*graphql.DocumentPage
		ExpectedErr     []error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListByBundleIDs", contextParam, bundleIDs, first, after).Return(docPages, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleToGraphQL", docsFirstBundle).Return(gqlDocsFirstBundle).Once()
				conv.On("MultipleToGraphQL", docsSecondBundle).Return(gqlDocsSecondBundle).Once()
				return conv
			},
			ExpectedResult: gqlDocPages,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when document listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListByBundleIDs", contextParam, bundleIDs, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)

			firstBundleParams := dataloader.ParamDocument{ID: firstBundleID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			secondBundleParams := dataloader.ParamDocument{ID: secondBundleID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			keys := []dataloader.ParamDocument{firstBundleParams, secondBundleParams}
			resolver := bundle.NewResolver(transact, nil, nil, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil)

			// WHEN
			result, err := resolver.DocumentsDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}

	t.Run("Returns error when there are no Bundles", func(t *testing.T) {
		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.DocumentsDataLoader([]dataloader.ParamDocument{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No Bundles found").Error())
	})

	t.Run("Returns error when start cursor is nil", func(t *testing.T) {
		params := dataloader.ParamDocument{ID: firstBundleID, Ctx: context.TODO(), First: nil, After: &gqlAfter}
		keys := []dataloader.ParamDocument{params}

		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.DocumentsDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInvalidDataError("missing required parameter 'first'").Error())
	})
}

func TestResolver_AddBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	desc := "bar"
	name := "baz"

	modelBundle := fixBundleModel(name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)
	gqlBundleInput := fixGQLBundleCreateInput(name, desc)
	modelBundleInput := fixModelBundleCreateInput(name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		BundleSvcFn     func() *automock.BundleService
		AppSvcFn        func() *automock.ApplicationService
		ConverterFn     func() *automock.BundleConverter
		ExpectedBundle  *graphql.Bundle
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			BundleSvcFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appID, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, extractTargetURLFromJSONArray(modelBundleInput.APIDefinitions[0].TargetURLs)).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: gqlBundle,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			BundleSvcFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when adding Bundle failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BundleSvcFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appID, modelBundleInput).Return("", testErr).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BundleSvcFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appID, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when updating base url failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BundleSvcFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appID, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, extractTargetURLFromJSONArray(modelBundleInput.APIDefinitions[0].TargetURLs)).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			BundleSvcFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appID, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, extractTargetURLFromJSONArray(modelBundleInput.APIDefinitions[0].TargetURLs)).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			BundleSvcFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appID, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("UpdateBaseURL", txtest.CtxWithDBMatcher(), appID, extractTargetURLFromJSONArray(modelBundleInput.APIDefinitions[0].TargetURLs)).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(nil, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			bundleSvc := testCase.BundleSvcFn()
			appSvc := testCase.AppSvcFn()
			converter := testCase.ConverterFn()

			resolver := bundle.NewResolver(transact, bundleSvc, nil, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil, appSvc)

			// WHEN
			result, err := resolver.AddBundle(context.TODO(), appID, gqlBundleInput)

			// then
			assert.Equal(t, testCase.ExpectedBundle, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			bundleSvc.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "bar"
	gqlBundleUpdateInput := fixGQLBundleUpdateInput(name, desc)
	modelBundleUpdateInput := fixModelBundleUpdateInput(name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)
	modelBundle := fixBundleModel(name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		ConverterFn     func() *automock.BundleConverter
		InputBundle     graphql.BundleUpdateInput
		ExpectedBundle  *graphql.Bundle
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: gqlBundle,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting from GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&model.BundleUpdateInput{}, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(nil, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := bundle.NewResolver(transact, svc, nil, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.UpdateBundle(context.TODO(), id, gqlBundleUpdateInput)

			// then
			assert.Equal(t, testCase.ExpectedBundle, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "desc"
	modelBundle := fixBundleModel(name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		APIDefFn        func() *automock.APIService
		EventDefFn      func() *automock.EventService
		ConverterFn     func() *automock.BundleConverter
		ExpectedBundle  *graphql.Bundle
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: gqlBundle,
			ExpectedErr:    nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when APIs deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Events deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			APIDefFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return apiSvc
			},
			EventDefFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("DeleteAllByBundleID", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return eventSvc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(nil, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			apiSvc := testCase.APIDefFn()
			eventSvc := testCase.EventDefFn()
			converter := testCase.ConverterFn()

			resolver := bundle.NewResolver(transact, svc, nil, nil, apiSvc, eventSvc, nil, converter, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.DeleteBundle(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedBundle, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
		})
	}
}

func TestResolver_InstanceAuth(t *testing.T) {
	// GIVEN
	id := "foo"
	modelBundleInstanceAuth := fixModelBundleInstanceAuth(id)
	gqlBundleInstanceAuth := fixGQLBundleInstanceAuth(id)
	bndl := fixGQLBundle("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn                  func() *automock.BundleInstanceAuthService
		ConverterFn                func() *automock.BundleInstanceAuthConverter
		InputID                    string
		Bundle                     *graphql.Bundle
		ExpectedBundleInstanceAuth *graphql.BundleInstanceAuth
		ExpectedErr                error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundleInstanceAuth, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("ToGraphQL", modelBundleInstanceAuth).Return(gqlBundleInstanceAuth, nil).Once()
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: gqlBundleInstanceAuth,
			ExpectedErr:                nil,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
		{
			Name:            "Returns error when conversion to graphql fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundleInstanceAuth, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("ToGraphQL", modelBundleInstanceAuth).Return(nil, testErr).Once()
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
		{
			Name:            "Returns nil when bundle instance auth for bundle not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundleInstanceAuth, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("ToGraphQL", modelBundleInstanceAuth).Return(gqlBundleInstanceAuth, nil).Once()
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := bundle.NewResolver(transact, nil, svc, nil, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.InstanceAuth(context.TODO(), testCase.Bundle, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedBundleInstanceAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Bundle is nil", func(t *testing.T) {
		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.InstanceAuth(context.TODO(), nil, "")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, apperrors.NewInternalError("Bundle cannot be empty").Error())
	})
}

func TestResolver_InstanceAuths(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	bndl := fixGQLBundle(bundleID, "foo", "bar")
	modelBundleInstanceAuths := []*model.BundleInstanceAuth{
		fixModelBundleInstanceAuth("foo"),
		fixModelBundleInstanceAuth("bar"),
	}

	gqlBundleInstanceAuths := []*graphql.BundleInstanceAuth{
		fixGQLBundleInstanceAuth("foo"),
		fixGQLBundleInstanceAuth("bar"),
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleInstanceAuthService
		ConverterFn     func() *automock.BundleInstanceAuthConverter
		ExpectedResult  []*graphql.BundleInstanceAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), bundleID).Return(modelBundleInstanceAuths, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("MultipleToGraphQL", modelBundleInstanceAuths).Return(gqlBundleInstanceAuths, nil).Once()
				return conv
			},
			ExpectedResult: gqlBundleInstanceAuths,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle Instance Auths listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), bundleID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle Instance Auths conversion to graphql failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), bundleID).Return(modelBundleInstanceAuths, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("MultipleToGraphQL", modelBundleInstanceAuths).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), bundleID).Return(modelBundleInstanceAuths, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("MultipleToGraphQL", modelBundleInstanceAuths).Return(gqlBundleInstanceAuths, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := bundle.NewResolver(transact, nil, svc, nil, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil)
			// WHEN
			result, err := resolver.InstanceAuths(context.TODO(), bndl)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Bundle is nil", func(t *testing.T) {
		resolver := bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.InstanceAuths(context.TODO(), nil)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, apperrors.NewInternalError("Bundle cannot be empty").Error())
	})
}

func extractTargetURLFromJSONArray(jsonTargetURL json.RawMessage) string {
	strTargetURL := string(jsonTargetURL)
	strTargetURL = strings.TrimPrefix(strTargetURL, `["`)
	strTargetURL = strings.TrimSuffix(strTargetURL, `"]`)

	return strTargetURL
}
