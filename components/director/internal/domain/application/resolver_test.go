package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
)

func TestResolver_CreateApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name                string
		TransactionerFn     func() *persistenceautomock.Transactioner
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		Input               graphql.ApplicationInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelApplication, nil).Once()
				svc.On("Create", ctxWithPersistenceTx, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application creation failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", ctxWithPersistenceTx, modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Returns error when application creation failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", ctxWithPersistenceTx, modelInput).Return("foo", nil).Once()
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			transactioner := testCase.TransactionerFn()
			resolver := application.NewResolver(transactioner, appCtx, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.CreateApplication(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactioner.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.ApplicationInput{
		Name:        "Foo",
		Description: &desc,
	}
	applicationID := "foo"

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name                string
		TransactionerFn     func() *persistenceautomock.Transactioner
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		ApplicationID       string
		Input               graphql.ApplicationInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelApplication, nil).Once()
				svc.On("Update", ctxWithPersistenceTx, applicationID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application update failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", ctxWithPersistenceTx, applicationID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Returns error when application retrieval failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", ctxWithPersistenceTx, applicationID, modelInput).Return(nil).Once()
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			transactioner := testCase.TransactionerFn()

			resolver := application.NewResolver(transactioner, appCtx, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.UpdateApplication(context.TODO(), testCase.ApplicationID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactioner.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name                string
		TransactionerFn     func() *persistenceautomock.Transactioner
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelApplication, nil).Once()
				svc.On("Delete", ctxWithPersistenceTx, "foo").Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application deletion failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(modelApplication, nil).Once()
				svc.On("Delete", ctxWithPersistenceTx, "foo").Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Returns error when application retrieval failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", ctxWithPersistenceTx, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			transactioner := testCase.TransactionerFn()

			resolver := application.NewResolver(transactioner, appCtx, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteApplication(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Application(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name                string
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(modelApplication, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when application retrieval failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", context.TODO(), "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputID:             "foo",
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.Application(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Applications(t *testing.T) {
	// given
	modelApplications := []*model.Application{
		fixModelApplication("foo", "Foo", "Lorem Ipsum"),
		fixModelApplication("bar", "Bar", "Lorem Ipsum"),
	}

	gqlApplications := []*graphql.Application{
		fixGQLApplication("foo", "Foo", "Lorem Ipsum"),
		fixGQLApplication("bar", "Bar", "Lorem Ipsum"),
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	query := "foo"
	filter := []*labelfilter.LabelFilter{
		{Key: "", Query: &query},
	}
	gqlFilter := []*graphql.LabelFilter{
		{Key: "", Query: &query},
	}
	testErr := errors.New("Test error")

	testCases := []struct {
		Name              string
		ServiceFn         func() *automock.ApplicationService
		ConverterFn       func() *automock.ApplicationConverter
		InputLabelFilters []*graphql.LabelFilter
		InputFirst        *int
		InputAfter        *graphql.PageCursor
		ExpectedResult    *graphql.ApplicationPage
		ExpectedErr       error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", context.TODO(), filter, &first, &after).Return(fixApplicationPage(modelApplications), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("MultipleToGraphQL", modelApplications).Return(gqlApplications).Once()
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    fixGQLApplicationPage(gqlApplications),
			ExpectedErr:       nil,
		},
		{
			Name: "Returns error when application listing failed",
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", context.TODO(), filter, &first, &after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.Applications(context.TODO(), testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_SetApplicationLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	gqlLabel := &graphql.Label{
		Key:   "key",
		Value: []string{"foo", "bar"},
	}
	modelLabel := &model.LabelInput{
		Key:        "key",
		Value:      []string{"foo", "bar"},
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name               string
		TransactionerFn    func() *persistenceautomock.Transactioner
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		InputValue         interface{}
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("SetLabel", ctxWithPersistenceTx, modelLabel).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValue:         gqlLabel.Value,
			ExpectedLabel:      gqlLabel,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when adding label to application failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("SetLabel", ctxWithPersistenceTx, modelLabel).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			InputValue:         gqlLabel.Value,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			transactioner := testCase.TransactionerFn()

			resolver := application.NewResolver(transactioner, appCtx, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.SetApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteApplicationLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"

	labelKey := "key"

	gqlLabel := &graphql.Label{
		Key:   labelKey,
		Value: []string{"foo", "bar"},
	}

	modelLabel := &model.Label{
		ID:         "b39ba24d-87fe-43fe-ac55-7f2e5ee04bcb",
		Tenant:     "tnt",
		Key:        labelKey,
		Value:      []string{"foo", "bar"},
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	}

	persistTx := &persistenceautomock.PersistenceTx{}
	persistTx.On("Commit").Return(nil)

	ctx := context.TODO()
	ctxWithPersistenceTx := context.WithValue(ctx, persistence.PersistenceCtxKey, persistTx)

	appCtx := &automock.ContextValueSetter{}
	appCtx.On("WithValue", ctx, persistence.PersistenceCtxKey, persistTx).Return(ctxWithPersistenceTx)

	testCases := []struct {
		Name               string
		TransactionerFn    func() *persistenceautomock.Transactioner
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name: "Success",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", ctxWithPersistenceTx, applicationID, labelKey).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", ctxWithPersistenceTx, applicationID, labelKey).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			ExpectedLabel:      gqlLabel,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when label retrieval failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", ctxWithPersistenceTx, applicationID, labelKey).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Returns error when deleting application's label failed",
			TransactionerFn: func() *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", ctxWithPersistenceTx, applicationID, labelKey).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", ctxWithPersistenceTx, applicationID, labelKey).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputApplicationID: applicationID,
			InputKey:           gqlLabel.Key,
			ExpectedLabel:      nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			transactioner := testCase.TransactionerFn()

			resolver := application.NewResolver(transactioner, appCtx, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactioner.AssertExpectations(t)
		})
	}
}

func TestResolver_Documents(t *testing.T) {
	// given
	applicationID := "fooid"
	modelDocuments := []*model.Document{
		fixModelDocument(applicationID, "foo"),
		fixModelDocument(applicationID, "bar"),
	}
	gqlDocuments := []*graphql.Document{
		fixGQLDocument("foo"),
		fixGQLDocument("bar"),
	}
	app := fixGQLApplication(applicationID, "foo", "bar")

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.DocumentService
		ConverterFn    func() *automock.DocumentConverter
		InputFirst     *int
		InputAfter     *graphql.PageCursor
		ExpectedResult *graphql.DocumentPage
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(fixModelDocumentPage(modelDocuments), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleToGraphQL", modelDocuments).Return(gqlDocuments).Once()
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: fixGQLDocumentPage(gqlDocuments),
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when document listing failed",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, nil, nil, svc, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.Documents(context.TODO(), app, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Webhooks(t *testing.T) {
	// given
	applicationID := "fooid"
	modelWebhooks := []*model.Webhook{
		fixModelWebhook(applicationID, "foo"),
		fixModelWebhook(applicationID, "bar"),
	}
	gqlWebhooks := []*graphql.Webhook{
		fixGQLWebhook("foo"),
		fixGQLWebhook("bar"),
	}
	app := fixGQLApplication(applicationID, "foo", "bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.WebhookService
		ConverterFn    func() *automock.WebhookConverter
		ExpectedResult []*graphql.Webhook
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("List", context.TODO(), applicationID).Return(modelWebhooks, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks).Once()
				return conv
			},
			ExpectedResult: gqlWebhooks,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when webhook listing failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("List", context.TODO(), applicationID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, svc, nil, nil, converter, nil, nil)

			// when
			result, err := resolver.Webhooks(context.TODO(), app)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Apis(t *testing.T) {
	// given
	applicationID := "1"
	group := "group"
	app := fixGQLApplication(applicationID, "foo", "bar")
	modelAPIDefinitions := []*model.APIDefinition{

		fixModelAPIDefinition("foo", applicationID, "Foo", "Lorem Ipsum", group),
		fixModelAPIDefinition("bar", applicationID, "Bar", "Lorem Ipsum", group),
	}

	gqlAPIDefinitions := []*graphql.APIDefinition{
		fixGQLAPIDefinition("foo", applicationID, "Foo", "Lorem Ipsum", group),
		fixGQLAPIDefinition("bar", applicationID, "Bar", "Lorem Ipsum", group),
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.APIService
		ConverterFn    func() *automock.APIConverter
		InputFirst     *int
		InputAfter     *graphql.PageCursor
		ExpectedResult *graphql.APIDefinitionPage
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", modelAPIDefinitions).Return(gqlAPIDefinitions).Once()
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: fixGQLAPIDefinitionPage(gqlAPIDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when APIS listing failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, svc, nil, nil, nil, nil, nil, nil, converter, nil)
			// when
			result, err := resolver.Apis(context.TODO(), app, &group, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_EventAPIs(t *testing.T) {
	// given
	applicationID := "1"
	group := "group"
	app := fixGQLApplication(applicationID, "foo", "bar")
	modelEventAPIDefinitions := []*model.EventAPIDefinition{

		fixModelEventAPIDefinition("foo", applicationID, "Foo", "Lorem Ipsum", group),
		fixModelEventAPIDefinition("bar", applicationID, "Bar", "Lorem Ipsum", group),
	}

	gqlEventAPIDefinitions := []*graphql.EventAPIDefinition{
		fixGQLEventAPIDefinition("foo", applicationID, "Foo", "Lorem Ipsum", group),
		fixGQLEventAPIDefinition("bar", applicationID, "Bar", "Lorem Ipsum", group),
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name           string
		ServiceFn      func() *automock.EventAPIService
		ConverterFn    func() *automock.EventAPIConverter
		InputFirst     *int
		InputAfter     *graphql.PageCursor
		ExpectedResult *graphql.EventAPIDefinitionPage
		ExpectedErr    error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(fixEventAPIDefinitionPage(modelEventAPIDefinitions), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("MultipleToGraphQL", modelEventAPIDefinitions).Return(gqlEventAPIDefinitions).Once()
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: fixGQLEventAPIDefinitionPage(gqlEventAPIDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when APIS listing failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("List", context.TODO(), applicationID, &first, &after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(nil, nil, nil, nil, svc, nil, nil, nil, nil, nil, nil, converter)
			// when
			result, err := resolver.EventAPIs(context.TODO(), app, &group, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_ApplicationsForRuntime(t *testing.T) {
	ctx := context.TODO()
	testError := errors.New("test error")

	modelApplications := []*model.Application{
		fixModelApplication("id1", "name", "desc"),
		fixModelApplication("id2", "name", "desc"),
	}

	applicationGraphQL := []*graphql.Application{
		fixGQLApplication("id1", "name", "desc"),
		fixGQLApplication("id2", "name", "desc"),
	}

	first := 10
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	runtimeID := "foo"
	testCases := []struct {
		Name           string
		AppConverterFn func() *automock.ApplicationConverter
		AppServiceFn   func() *automock.ApplicationService
		InputRuntimeID string
		InputFirst     *int
		InputAfter     *graphql.PageCursor
		ExpectedResult *graphql.ApplicationPage
		ExpectedError  error
	}{
		{
			Name: "Success",
			AppServiceFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListByRuntimeID", ctx, runtimeID, &first, &after).Return(fixApplicationPage(modelApplications), nil).Once()
				return appService
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				appConverter.On("MultipleToGraphQL", modelApplications).Return(applicationGraphQL).Once()
				return appConverter
			},
			InputRuntimeID: runtimeID,
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: fixGQLApplicationPage(applicationGraphQL),
			ExpectedError:  nil,
		},
		{
			Name: "Returns error when application listing failed",
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListByRuntimeID", ctx, runtimeID, &first, &after).Return(nil, testError).Once()
				return appSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			InputRuntimeID: runtimeID,
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: nil,
			ExpectedError:  testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			applicationSvc := testCase.AppServiceFn()
			applicationConverter := testCase.AppConverterFn()

			resolver := application.NewResolver(nil, nil, applicationSvc, nil, nil, nil, nil, applicationConverter, nil, nil, nil, nil)

			//WHEN
			result, err := resolver.ApplicationsForRuntime(ctx, testCase.InputRuntimeID, testCase.InputFirst, testCase.InputAfter)

			//THEN
			if testCase.ExpectedError != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedResult, result)
			applicationSvc.AssertExpectations(t)
			applicationConverter.AssertExpectations(t)

		})
	}
}
