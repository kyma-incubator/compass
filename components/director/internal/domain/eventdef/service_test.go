package eventdef_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	event "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"

	eventDefinition := fixEventDefinitionModel(id, name)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		Input              model.EventDefinitionInput
		InputID            string
		ExpectedDocument   *model.EventDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   eventDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when EventDefinition retrieval failed",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   eventDefinition,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := event.NewService(repo, nil, nil, nil)

			// when
			document, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDocument, document)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetForBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	bndlID := "foobar"
	name := "foo"

	eventDefinition := fixEventDefinitionModel(id, name)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		Input              model.EventDefinitionInput
		InputID            string
		BundleID           string
		ExpectedEvent      *model.EventDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetForBundle", ctx, tenantID, id, bndlID).Return(eventDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			BundleID:           bndlID,
			ExpectedEvent:      eventDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when EventDefinition retrieval failed",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetForBundle", ctx, tenantID, id, bndlID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			BundleID:           bndlID,
			ExpectedEvent:      nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := event.NewService(repo, nil, nil, nil)

			// when
			eventDef, err := svc.GetForBundle(ctx, testCase.InputID, testCase.BundleID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedEvent, eventDef)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetForBundle(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByBundleIDs(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	firstEventID := "foo"
	secondEventID := "foo2"
	firstBundleID := "bar"
	secondBundleID := "bar2"
	name := "foo"
	numberOfEventsInFirstBundle := 1
	numberOfEventsInSecondBundle := 1
	bundleIDs := []string{firstBundleID, secondBundleID}

	eventFirstBundle := fixEventDefinitionModel(firstEventID, name)
	eventSecondBundle := fixEventDefinitionModel(secondEventID, name)

	eventFirstBundleReference := fixModelBundleReference(firstBundleID, firstEventID)
	eventSecondBundleReference := fixModelBundleReference(secondBundleID, secondEventID)
	bundleRefs := []*model.BundleReference{eventFirstBundleReference, eventSecondBundleReference}
	totalCounts := map[string]int{firstBundleID: numberOfEventsInFirstBundle, secondBundleID: numberOfEventsInSecondBundle}

	eventsFirstBundle := []*model.EventDefinition{eventFirstBundle}
	eventsSecondBundle := []*model.EventDefinition{eventSecondBundle}

	eventPageFirstBundle := &model.EventDefinitionPage{
		Data:       eventsFirstBundle,
		TotalCount: len(eventsFirstBundle),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}
	eventPageSecondBundle := &model.EventDefinitionPage{
		Data:       eventsSecondBundle,
		TotalCount: len(eventsSecondBundle),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	eventPages := []*model.EventDefinitionPage{eventPageFirstBundle, eventPageSecondBundle}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.EventAPIRepository
		BundleRefSvcFn     func() *automock.BundleReferenceService
		ExpectedResult     []*model.EventDefinitionPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", ctx, model.BundleEventReference, bundleIDs, 2, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("ListByBundleIDs", ctx, tenantID, bundleIDs, bundleRefs, totalCounts, 2, after).Return(eventPages, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     eventPages,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				return svc
			},
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     eventPages,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				return svc
			},
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     eventPages,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when EventDefinition BundleReferences listing failed",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", ctx, model.BundleEventReference, bundleIDs, 2, after).Return(nil, nil, testErr).Once()
				return svc
			},
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			PageSize:           2,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when EventDefinition listing failed",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", ctx, model.BundleEventReference, bundleIDs, 2, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("ListByBundleIDs", ctx, tenantID, bundleIDs, bundleRefs, totalCounts, 2, after).Return(nil, testErr).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			bndlRefSvc := testCase.BundleRefSvcFn()

			svc := event.NewService(repo, nil, nil, bndlRefSvc)

			// when
			events, err := svc.ListByBundleIDs(ctx, bundleIDs, testCase.PageSize, after)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, events)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByBundleIDs(context.TODO(), nil, 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"

	apiDefinitions := []*model.EventDefinition{
		fixEventDefinitionModel(id, name),
		fixEventDefinitionModel(id, name),
		fixEventDefinitionModel(id, name),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		ExpectedResult     []*model.EventDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(apiDefinitions, nil).Once()
				return repo
			},
			ExpectedResult:     apiDefinitions,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when EventDefinition listing failed",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := event.NewService(repo, nil, nil, nil)

			// when
			docs, err := svc.ListByApplicationID(ctx, appID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	bundleID := "bndlid"
	bundleID2 := "bndlid2"
	packageID := packageID
	name := "Foo"

	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "test"
	spec2 := "test2"

	modelInput := model.EventDefinitionInput{
		Name:         name,
		VersionInput: &model.VersionInput{},
	}

	modelSpecsInput := []*model.SpecInput{
		{
			Data: &spec,
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
		{
			Data: &spec2,
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
	}

	modelEventDefinition := &model.EventDefinition{
		PackageID:     &packageID,
		ApplicationID: appID,
		Name:          name,
		Version:       &model.Version{},
		BaseEntity: &model.BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	bundleReferenceInput := &model.BundleReferenceInput{}
	bundleIDs := []string{bundleID, bundleID2}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name              string
		RepositoryFn      func() *automock.EventAPIRepository
		UIDServiceFn      func() *automock.UIDService
		SpecServiceFn     func() *automock.SpecService
		BundleReferenceFn func() *automock.BundleReferenceService
		Input             model.EventDefinitionInput
		SpecsInput        []*model.SpecInput
		BundleIDs         []string
		ExpectedErr       error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.EventSpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.EventSpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleEventReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				return svc
			},
			Input:      modelInput,
			SpecsInput: modelSpecsInput,
		},
		{
			Name: "Success in ORD scenario where many bundle ids are passed",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.EventSpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.EventSpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleEventReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleEventReference, str.Ptr(id), str.Ptr(bundleID2)).Return(nil).Once()
				return svc
			},
			Input:      modelInput,
			SpecsInput: modelSpecsInput,
			BundleIDs:  bundleIDs,
		},
		{
			Name: "Error - Event Creation",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventDefinition).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			Input:       modelInput,
			SpecsInput:  modelSpecsInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Spec Creation",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.EventSpecReference, id).Return("", testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			Input:       modelInput,
			SpecsInput:  modelSpecsInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - BundleReference Event Creation",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.EventSpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.EventSpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleEventReference, str.Ptr(id), str.Ptr(bundleID)).Return(testErr).Once()
				return svc
			},
			Input:       modelInput,
			SpecsInput:  modelSpecsInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error in ORD scenario - BundleReference Event Creation",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.EventSpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.EventSpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleEventReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleEventReference, str.Ptr(id), str.Ptr(bundleID2)).Return(testErr).Once()
				return svc
			},
			Input:       modelInput,
			SpecsInput:  modelSpecsInput,
			BundleIDs:   bundleIDs,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()
			specService := testCase.SpecServiceFn()
			bundleReferenceService := testCase.BundleReferenceFn()

			svc := event.NewService(repo, uidService, specService, bundleReferenceService)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			result, err := svc.Create(ctx, appID, &bundleID, &packageID, testCase.Input, testCase.SpecsInput, testCase.BundleIDs, 0)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			repo.AssertExpectations(t)
			specService.AssertExpectations(t)
			uidService.AssertExpectations(t)
			bundleReferenceService.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", nil, nil, model.EventDefinitionInput{}, []*model.SpecInput{}, []string{}, 0)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "spec"

	modelInput := model.EventDefinitionInput{
		Name:         "Foo",
		VersionInput: &model.VersionInput{},
	}

	modelSpecInput := model.SpecInput{
		Data: &spec,
		FetchRequest: &model.FetchRequestInput{
			URL: frURL,
		},
	}

	modelSpec := &model.Spec{
		ID:         id,
		ObjectType: model.EventSpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputEventDefinitionModel := mock.MatchedBy(func(event *model.EventDefinition) bool {
		return event.Name == modelInput.Name
	})

	eventDefinitionModel := &model.EventDefinition{
		Name:    "Bar",
		Version: &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name          string
		RepositoryFn  func() *automock.EventAPIRepository
		SpecServiceFn func() *automock.SpecService
		Input         model.EventDefinitionInput
		InputID       string
		SpecInput     *model.SpecInput
		ExpectedErr   error
	}{
		{
			Name: "Success When Spec is Found should update it",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.EventSpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, model.EventSpecReference, id).Return(nil).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success When Spec is not found should create in",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.EventSpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, model.EventSpecReference, id).Return("id", nil).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(testErr).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Spec Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.EventSpecReference, id).Return(nil, testErr).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Spec Creation Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.EventSpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, model.EventSpecReference, id).Return("", testErr).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Spec Update Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.EventSpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, model.EventSpecReference, id).Return(testErr).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			specSvc := testCase.SpecServiceFn()

			svc := event.NewService(repo, nil, specSvc, nil)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input, testCase.SpecInput)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			specSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", model.EventDefinitionInput{}, &model.SpecInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateManyBundles(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	bndlID1 := "id1"
	bndlID2 := "id2"
	bndlID3 := "id3"
	bndlID4 := "id4"
	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "spec"

	modelInput := model.EventDefinitionInput{
		Name:         "Foo",
		VersionInput: &model.VersionInput{},
	}

	modelSpecInput := model.SpecInput{
		Data: &spec,
		FetchRequest: &model.FetchRequestInput{
			URL: frURL,
		},
	}

	modelSpec := &model.Spec{
		ID:         id,
		ObjectType: model.EventSpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputEventDefinitionModel := mock.MatchedBy(func(event *model.EventDefinition) bool {
		return event.Name == modelInput.Name
	})

	eventDefinitionModel := &model.EventDefinition{
		Name:    "Bar",
		Version: &model.Version{},
	}

	bundleIDsForCreation := []string{bndlID1, bndlID2}
	bundleIDsForDeletion := []string{bndlID3, bndlID4}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                     string
		RepositoryFn             func() *automock.EventAPIRepository
		SpecServiceFn            func() *automock.SpecService
		BundleReferenceServiceFn func() *automock.BundleReferenceService
		Input                    model.EventDefinitionInput
		InputID                  string
		SpecInput                *model.SpecInput
		BundleIDsForCreation     []string
		BundleIDsForDeletion     []string
		ExpectedErr              error
	}{
		{
			Name: "Success in ORD case",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.EventSpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, model.EventSpecReference, id).Return(nil).Once()
				return svc
			},
			BundleReferenceServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, model.BundleReferenceInput{}, model.BundleEventReference, &id, &bndlID1).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, model.BundleReferenceInput{}, model.BundleEventReference, &id, &bndlID2).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctx, model.BundleEventReference, &id, &bndlID3).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctx, model.BundleEventReference, &id, &bndlID4).Return(nil).Once()
				return svc
			},
			InputID:              "foo",
			Input:                modelInput,
			SpecInput:            &modelSpecInput,
			BundleIDsForCreation: bundleIDsForCreation,
			BundleIDsForDeletion: bundleIDsForDeletion,
			ExpectedErr:          nil,
		},
		{
			Name: "Error on BundleReference creation",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, model.BundleReferenceInput{}, model.BundleEventReference, &id, &bndlID1).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, model.BundleReferenceInput{}, model.BundleEventReference, &id, &bndlID2).Return(testErr).Once()
				return svc
			},
			InputID:              "foo",
			Input:                modelInput,
			SpecInput:            &modelSpecInput,
			BundleIDsForCreation: bundleIDsForCreation,
			BundleIDsForDeletion: bundleIDsForDeletion,
			ExpectedErr:          testErr,
		},
		{
			Name: "Error on BundleReference deletion",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceServiceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, model.BundleReferenceInput{}, model.BundleEventReference, &id, &bndlID1).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, model.BundleReferenceInput{}, model.BundleEventReference, &id, &bndlID2).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctx, model.BundleEventReference, &id, &bndlID3).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctx, model.BundleEventReference, &id, &bndlID4).Return(testErr).Once()
				return svc
			},
			InputID:              "foo",
			Input:                modelInput,
			SpecInput:            &modelSpecInput,
			BundleIDsForCreation: bundleIDsForCreation,
			BundleIDsForDeletion: bundleIDsForDeletion,
			ExpectedErr:          testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			specSvc := testCase.SpecServiceFn()
			bndlRefSvc := testCase.BundleReferenceServiceFn()

			svc := event.NewService(repo, nil, specSvc, bndlRefSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.UpdateInManyBundles(ctx, testCase.InputID, testCase.Input, testCase.SpecInput, testCase.BundleIDsForCreation, testCase.BundleIDsForDeletion, 0)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			specSvc.AssertExpectations(t)
			bndlRefSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.UpdateInManyBundles(context.TODO(), "", model.EventDefinitionInput{}, &model.SpecInput{}, []string{}, []string{}, 0)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.EventAPIRepository
		Input        model.EventDefinitionInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := event.NewService(repo, nil, nil, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_DeleteAllByBundleID(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	id := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.EventAPIRepository
		Input        model.EventDefinitionInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("DeleteAllByBundleID", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("DeleteAllByBundleID", ctx, tenantID, id).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := event.NewService(repo, nil, nil, nil)

			// when
			err := svc.DeleteAllByBundleID(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListFetchRequests(t *testing.T) {
	// given
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testErr := errors.New("Test error")

	frURL := "foo.bar"
	firstFRID := "frID"
	secondFRID := "frID2"
	firstSpecID := "specID"
	secondSpecID := "specID2"
	specIDs := []string{firstSpecID, secondSpecID}
	timestamp := time.Now()

	firstFetchRequest := fixModelFetchRequest(firstFRID, frURL, timestamp)
	secondFetchRequest := fixModelFetchRequest(secondFRID, frURL, timestamp)
	fetchRequests := []*model.FetchRequest{firstFetchRequest, secondFetchRequest}

	testCases := []struct {
		Name                  string
		SpecServiceFn         func() *automock.SpecService
		ExpectedFetchRequests []*model.FetchRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "Success",
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("ListFetchRequestsByReferenceObjectIDs", ctx, tenantID, specIDs).Return(fetchRequests, nil).Once()
				return svc
			},
			ExpectedFetchRequests: fetchRequests,
		},
		{
			Name: "Success - Fetch Request Not Found",
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("ListFetchRequestsByReferenceObjectIDs", ctx, tenantID, specIDs).Return(nil, apperrors.NewNotFoundError(resource.FetchRequest, "")).Once()
				return svc
			},
			ExpectedFetchRequests: nil,
		},
		{
			Name: "Error while listing Fetch Requests",
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("ListFetchRequestsByReferenceObjectIDs", ctx, tenantID, specIDs).Return(nil, testErr).Once()
				return svc
			},
			ExpectedFetchRequests: nil,
			ExpectedErrMessage:    testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			specService := testCase.SpecServiceFn()

			svc := event.NewService(nil, nil, specService, nil)

			// when
			frs, err := svc.ListFetchRequests(ctx, specIDs)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, frs, testCase.ExpectedFetchRequests)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			specService.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := event.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.ListFetchRequests(context.TODO(), nil)
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}
