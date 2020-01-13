package application_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var contextParam = txtest.CtxWithDBMatcher()

func TestResolver_RegisterApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationRegisterInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.ApplicationRegisterInput{
		Name:        "Foo",
		Description: &desc,
	}
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		Input               graphql.ApplicationRegisterInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(modelApplication, nil).Once()
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(modelApplication, nil).Once()
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", contextParam, modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.RegisterApplication(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateApplication(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlApplication := fixGQLApplication("foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.ApplicationUpdateInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.ApplicationUpdateInput{
		Name:        "Foo",
		Description: &desc,
	}
	applicationID := "foo"

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		ApplicationID       string
		Input               graphql.ApplicationUpdateInput
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(modelApplication, nil).Once()
				svc.On("Update", contextParam, applicationID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(modelApplication, nil).Once()
				svc.On("Update", contextParam, applicationID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", contextParam, applicationID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Update", contextParam, applicationID, modelInput).Return(nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ApplicationID:       applicationID,
			Input:               gqlInput,
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
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
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.UpdateApplication(context.TODO(), testCase.ApplicationID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_UnregisterApplication(t *testing.T) {
	// given
	appID := uuid.New()
	modelApplication := fixModelApplication(appID.String(), "tenant-foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication(appID.String(), "Foo", "Bar")
	testErr := errors.New("Test error")
	testAuths := fixOAuths()
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		EventingSvcFn       func() *automock.EventingService
		SysAuthServiceFn    func() *automock.SystemAuthService
		OAuth20ServiceFn    func() *automock.OAuth20Service
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, appID.String()).Return(modelApplication, nil).Once()
				svc.On("Delete", contextParam, appID.String()).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("ToGraphQL", modelApplication).Return(gqlApplication).Once()
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", contextParam, appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: gqlApplication,
			ExpectedErr:         nil,
		},
		{
			Name:            "Return error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, appID.String()).Return(modelApplication, nil).Once()
				svc.On("Delete", contextParam, appID.String()).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", contextParam, appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)

				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, appID.String()).Return(modelApplication, nil).Once()
				svc.On("Delete", contextParam, appID.String()).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", contextParam, appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)

				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, appID.String()).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when transaction starting failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when listing all auths failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, appID.String()).Return(modelApplication, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", contextParam, appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.ApplicationReference, modelApplication.ID).Return(nil, testErr)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Return error when removing oauth from hydra",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, appID.String()).Return(modelApplication, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", contextParam, appID).Return(nil, nil).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.ApplicationReference, modelApplication.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(testErr)
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		}, {
			Name:            "Returns error when removing default eventing labels",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, appID.String()).Return(modelApplication, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			EventingSvcFn: func() *automock.EventingService {
				svc := &automock.EventingService{}
				svc.On("CleanupAfterUnregisteringApplication", contextParam, appID).Return(nil, testErr).Once()
				return svc
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			InputID:             appID.String(),
			ExpectedApplication: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			eventingSvc := testCase.EventingSvcFn()
			persistTx, transact := testCase.TransactionerFn()
			sysAuthSvc := testCase.SysAuthServiceFn()
			oAuth20Svc := testCase.OAuth20ServiceFn()
			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, oAuth20Svc, sysAuthSvc, nil, nil, nil, nil, nil, nil, eventingSvc)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.UnregisterApplication(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedApplication, result)
			if testCase.ExpectedErr != nil {
				assert.EqualError(t, testCase.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, svc, converter, persistTx, transact, sysAuthSvc, oAuth20Svc, eventingSvc)
		})
	}
}

func TestResolver_Application(t *testing.T) {
	// given
	modelApplication := fixModelApplication("foo", "tenant-foo", "Foo", "Bar")
	gqlApplication := fixGQLApplication("foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name                string
		PersistenceFn       func() *persistenceautomock.PersistenceTx
		TransactionerFn     func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn           func() *automock.ApplicationService
		ConverterFn         func() *automock.ApplicationConverter
		InputID             string
		ExpectedApplication *graphql.Application
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(modelApplication, nil).Once()

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
			Name:            "Returns error when application retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()

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
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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
		fixModelApplication("foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelApplication("bar", "tenant-bar", "Bar", "Lorem Ipsum"),
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
		PersistenceFn     func() *persistenceautomock.PersistenceTx
		TransactionerFn   func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn         func() *automock.ApplicationService
		ConverterFn       func() *automock.ApplicationConverter
		InputLabelFilters []*graphql.LabelFilter
		ExpectedResult    *graphql.ApplicationPage
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", contextParam, filter, first, after).Return(fixApplicationPage(modelApplications), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("MultipleToGraphQL", modelApplications).Return(gqlApplications).Once()
				return conv
			},
			InputLabelFilters: gqlFilter,
			ExpectedResult:    fixGQLApplicationPage(gqlApplications),
			ExpectedErr:       nil,
		},
		{
			Name:            "Returns error when application listing failed",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("List", contextParam, filter, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				return conv
			},
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.Applications(context.TODO(), testCase.InputLabelFilters, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_ApplicationsForRuntime(t *testing.T) {
	testError := errors.New("test error")

	modelApplications := []*model.Application{
		fixModelApplication("id1", "tenant-foo", "name", "desc"),
		fixModelApplication("id2", "tenant-bar", "name", "desc"),
	}

	applicationGraphQL := []*graphql.Application{
		fixGQLApplication("id1", "name", "desc"),
		fixGQLApplication("id2", "name", "desc"),
	}

	first := 10
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	txGen := txtest.NewTransactionContextGenerator(testError)

	runtimeUUID := uuid.New()
	runtimeID := runtimeUUID.String()
	testCases := []struct {
		Name            string
		AppConverterFn  func() *automock.ApplicationConverter
		AppServiceFn    func() *automock.ApplicationService
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		InputRuntimeID  string
		ExpectedResult  *graphql.ApplicationPage
		ExpectedError   error
	}{
		{
			Name: "Success",
			AppServiceFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListByRuntimeID", contextParam, runtimeUUID, first, after).Return(fixApplicationPage(modelApplications), nil).Once()
				return appService
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				appConverter.On("MultipleToGraphQL", modelApplications).Return(applicationGraphQL).Once()
				return appConverter
			},
			TransactionerFn: txGen.ThatSucceeds,
			InputRuntimeID:  runtimeID,
			ExpectedResult:  fixGQLApplicationPage(applicationGraphQL),
			ExpectedError:   nil,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			AppServiceFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListByRuntimeID", contextParam, runtimeUUID, first, after).Return(fixApplicationPage(modelApplications), nil).Once()
				return appService
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			InputRuntimeID: runtimeID,
			ExpectedResult: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Returns error when application listing failed",
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListByRuntimeID", contextParam, runtimeUUID, first, after).Return(nil, testError).Once()
				return appSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputRuntimeID:  runtimeID,
			ExpectedResult:  nil,
			ExpectedError:   testError,
		},
		{
			Name: "Returns error when starting transaction failed",
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			TransactionerFn: txGen.ThatFailsOnBegin,
			InputRuntimeID:  runtimeID,
			ExpectedResult:  nil,
			ExpectedError:   testError,
		},
		{
			Name: "Returns error when runtimeID is not UUID",
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConverter := &automock.ApplicationConverter{}
				return appConverter
			},
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			InputRuntimeID:  "blabla",
			ExpectedResult:  nil,
			ExpectedError:   errors.New("invalid UUID length"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			applicationSvc := testCase.AppServiceFn()
			applicationConverter := testCase.AppConverterFn()
			persistTx, transact := testCase.TransactionerFn()

			resolver := application.NewResolver(transact, applicationSvc, nil, nil, nil, nil, nil, nil, applicationConverter, nil, nil, nil, nil, nil, nil)

			//WHEN
			result, err := resolver.ApplicationsForRuntime(context.TODO(), testCase.InputRuntimeID, &first, &gqlAfter)

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
			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
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

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		InputValue         interface{}
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("SetLabel", contextParam, modelLabel).Return(nil).Once()
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
			Name:            "Returns error when adding label to application failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("SetLabel", contextParam, modelLabel).Return(testErr).Once()
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
			persistTx := testCase.PersistenceFn()
			transactioner := testCase.TransactionerFn(persistTx)

			resolver := application.NewResolver(transactioner, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.SetApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Label input validation failed", func(t *testing.T) {
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// when
		result, err := resolver.SetApplicationLabel(context.TODO(), "", "", "")

		// then
		require.Nil(t, result)
		require.Error(t, err)
		assert.EqualError(t, err, "validation error for type LabelInput: key: cannot be blank; value: cannot be blank.")
	})
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

	testCases := []struct {
		Name               string
		PersistenceFn      func() *persistenceautomock.PersistenceTx
		TransactionerFn    func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn          func() *automock.ApplicationService
		ConverterFn        func() *automock.ApplicationConverter
		InputApplicationID string
		InputKey           string
		ExpectedLabel      *graphql.Label
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", contextParam, applicationID, labelKey).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, applicationID, labelKey).Return(nil).Once()
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
			Name:            "Returns error when label retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", contextParam, applicationID, labelKey).Return(nil, testErr).Once()
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
			Name:            "Returns error when deleting application's label failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("GetLabel", contextParam, applicationID, labelKey).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, applicationID, labelKey).Return(testErr).Once()
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
			persistTx := testCase.PersistenceFn()
			transactioner := testCase.TransactionerFn(persistTx)

			resolver := application.NewResolver(transactioner, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteApplicationLabel(context.TODO(), testCase.InputApplicationID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transactioner.AssertExpectations(t)
			persistTx.AssertExpectations(t)
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
		Name            string
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.DocumentConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedResult  *graphql.DocumentPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("List", contextParam, applicationID, first, after).Return(fixModelDocumentPage(modelDocuments), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleToGraphQL", modelDocuments).Return(gqlDocuments).Once()
				return conv
			},
			ExpectedResult: fixGQLDocumentPage(gqlDocuments),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when document listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("List", contextParam, applicationID, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
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
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)

			resolver := application.NewResolver(transact, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil)

			// when
			result, err := resolver.Documents(context.TODO(), app, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
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
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		ExpectedResult  []*graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("List", contextParam, applicationID).Return(modelWebhooks, nil).Once()
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
			Name:            "Returns error when webhook listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("List", contextParam, applicationID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name: "Returns error on starting transaction",
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(nil, testErr).Once()
				return transact
			},
			PersistenceFn: txtest.PersistenceContextThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error on committing transaction",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(testErr).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("List", contextParam, applicationID).Return(modelWebhooks, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			mockPersistence := testCase.PersistenceFn()
			mockTransactioner := testCase.TransactionerFn(mockPersistence)

			resolver := application.NewResolver(mockTransactioner, nil, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil)

			// when
			result, err := resolver.Webhooks(context.TODO(), app)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			mockPersistence.AssertExpectations(t)
			mockTransactioner.AssertExpectations(t)
		})
	}
}

func TestResolver_Apis(t *testing.T) {
	// given
	testErr := errors.New("test error")

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

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.APIConverter
		ExpectedResult  *graphql.APIDefinitionPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("List", txtest.CtxWithDBMatcher(), applicationID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", modelAPIDefinitions).Return(gqlAPIDefinitions).Once()
				return conv
			},
			ExpectedResult: fixGQLAPIDefinitionPage(gqlAPIDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when APIS listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("List", txtest.CtxWithDBMatcher(), applicationID, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("List", txtest.CtxWithDBMatcher(), applicationID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, nil, svc, nil, nil, nil, nil, nil, nil, nil, nil, converter, nil, nil, nil)
			// when
			result, err := resolver.ApiDefinitions(context.TODO(), app, &group, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_EventAPIs(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "1"
	group := "group"
	app := fixGQLApplication(applicationID, "foo", "bar")
	modelEventAPIDefinitions := []*model.EventDefinition{

		fixModelEventAPIDefinition("foo", applicationID, "Foo", "Lorem Ipsum", group),
		fixModelEventAPIDefinition("bar", applicationID, "Bar", "Lorem Ipsum", group),
	}

	gqlEventAPIDefinitions := []*graphql.EventDefinition{
		fixGQLEventDefinition("foo", applicationID, "Foo", "Lorem Ipsum", group),
		fixGQLEventDefinition("bar", applicationID, "Bar", "Lorem Ipsum", group),
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventDefinitionService
		ConverterFn     func() *automock.EventAPIConverter
		InputFirst      *int
		InputAfter      *graphql.PageCursor
		ExpectedResult  *graphql.EventDefinitionPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefinitionService {
				svc := &automock.EventDefinitionService{}
				svc.On("List", contextParam, applicationID, first, after).Return(fixEventAPIDefinitionPage(modelEventAPIDefinitions), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("MultipleToGraphQL", modelEventAPIDefinitions).Return(gqlEventAPIDefinitions).Once()
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: fixGQLEventDefinitionPage(gqlEventAPIDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when APIS listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefinitionService {
				svc := &automock.EventDefinitionService{}
				svc.On("List", contextParam, applicationID, first, after).Return(nil, testErr).Once()
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
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, nil, nil, svc, nil, nil, nil, nil, nil, nil, nil, nil, converter, nil, nil)
			// when
			result, err := resolver.EventDefinitions(context.TODO(), app, &group, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_EventAPI(t *testing.T) {
	// given
	id := "bar"

	modelAPI := fixMinModelEventAPIDefinition(id, "placeholder")
	gqlAPI := fixGQLEventDefinition(id, "placeholder", "placeholder", "placeholder", "placeholder")
	app := fixGQLApplication("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventDefinitionService
		ConverterFn     func() *automock.EventAPIConverter
		InputID         string
		Application     *graphql.Application
		ExpectedAPI     *graphql.EventDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefinitionService {
				svc := &automock.EventDefinitionService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			InputID:     "foo",
			Application: app,
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefinitionService {
				svc := &automock.EventDefinitionService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			InputID:     "foo",
			Application: app,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns null when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefinitionService {
				svc := &automock.EventDefinitionService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError("")).Once()

				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			InputID:     "foo",
			Application: app,
			ExpectedAPI: nil,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefinitionService {
				svc := &automock.EventDefinitionService{}

				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			InputID:     "foo",
			Application: app,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefinitionService {
				svc := &automock.EventDefinitionService{}
				svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			InputID:     "foo",
			Application: app,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := application.NewResolver(transact, nil, nil, svc, nil, nil, nil, nil, nil, nil, nil, nil, converter, nil, nil)

			// when
			result, err := resolver.EventDefinition(context.TODO(), testCase.InputID, testCase.Application)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
func TestResolver_API(t *testing.T) {
	{
		// given
		id := "bar"
		appId := "1"
		modelAPI := fixModelAPIDefinition(id, appId, "name", "bar", "test")
		gqlAPI := fixGQLAPIDefinition(id, appId, "name", "bar", "test")
		app := fixGQLApplication("foo", "foo", "foo")
		testErr := errors.New("Test error")
		txGen := txtest.NewTransactionContextGenerator(testErr)

		testCases := []struct {
			Name            string
			TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ServiceFn       func() *automock.APIService
			ConverterFn     func() *automock.APIConverter
			InputID         string
			Application     *graphql.Application
			ExpectedAPI     *graphql.APIDefinition
			ExpectedErr     error
		}{
			{
				Name:            "Success",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
					return conv
				},
				InputID:     "foo",
				Application: app,
				ExpectedAPI: gqlAPI,
				ExpectedErr: nil,
			},
			{
				Name:            "Returns error when application retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Application: app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns null when application retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError("")).Once()

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Application: app,
				ExpectedAPI: nil,
				ExpectedErr: nil,
			},
			{
				Name:            "Returns error when commit begin error",
				TransactionerFn: txGen.ThatFailsOnBegin,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Application: app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when commit failed",
				TransactionerFn: txGen.ThatFailsOnCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForApplication", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Application: app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				persist, transact := testCase.TransactionerFn()
				svc := testCase.ServiceFn()
				converter := testCase.ConverterFn()

				resolver := application.NewResolver(transact, nil, svc, nil, nil, nil, nil, nil, nil, nil, nil, converter, nil, nil, nil)

				// when
				result, err := resolver.APIDefinition(context.TODO(), testCase.InputID, testCase.Application)

				// then
				assert.Equal(t, testCase.ExpectedAPI, result)
				assert.Equal(t, testCase.ExpectedErr, err)

				svc.AssertExpectations(t)
				persist.AssertExpectations(t)
				transact.AssertExpectations(t)
				converter.AssertExpectations(t)
			})
		}
	}
}

func TestResolver_Labels(t *testing.T) {
	// given

	id := "foo"
	tenant := "tenant"
	labelKey := "key"
	labelValue := "val"

	gqlApp := fixGQLApplication(id, "name", "desc")

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.ApplicationLabelableObject,
		},
		"def": {
			ID:         "def",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.ApplicationLabelableObject,
		},
	}

	gqlLabels := &graphql.Labels{
		labelKey: labelValue,
		labelKey: labelValue,
	}

	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.ApplicationService
		InputApp        *graphql.Application
		InputKey        string
		ExpectedResult  *graphql.Labels
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when label listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListLabels", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)

			resolver := application.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

			// when
			result, err := resolver.Labels(context.TODO(), gqlApp, &testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			transact.AssertExpectations(t)
			persistTx.AssertExpectations(t)
		})
	}
}

func TestResolver_Auths(t *testing.T) {
	// given
	id := "foo"
	testError := errors.New("error")
	gqlApp := fixGQLApplication(id, "name", "desc")
	txGen := txtest.NewTransactionContextGenerator(testError)

	sysAuthModels := []model.SystemAuth{{ID: "id1", AppID: &id}, {ID: "id2", AppID: &id}}
	sysAuthGQL := []*graphql.SystemAuth{{ID: "id1"}, {ID: "id2"}}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.SystemAuthService
		SysAuthConvFn   func() *automock.SystemAuthConverter
		InputApp        *graphql.Application
		ExpectedResult  []*graphql.SystemAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.ApplicationReference, id).Return(sysAuthModels, nil).Once()
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &sysAuthModels[0]).Return(sysAuthGQL[0]).Once()
				sysAuthConv.On("ToGraphQL", &sysAuthModels[1]).Return(sysAuthGQL[1]).Once()
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: sysAuthGQL,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.ApplicationReference, id).Return(sysAuthModels, nil).Once()
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when list for SystemAuths failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", txtest.CtxWithDBMatcher(), model.ApplicationReference, id).Return([]model.SystemAuth{}, testError).Once()
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			InputApp:       gqlApp,
			ExpectedResult: nil,
			ExpectedErr:    testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			persist, transact := testCase.TransactionerFn()
			conv := testCase.SysAuthConvFn()

			resolver := application.NewResolver(transact, nil, nil, nil, nil, nil, nil, svc, nil, nil, nil, nil, nil, conv, nil)

			// when
			result, err := resolver.Auths(context.TODO(), testCase.InputApp)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			conv.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
		})
	}

	t.Run("Returns error when application is nil", func(t *testing.T) {
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		//WHEN
		_, err := resolver.Auths(context.TODO(), nil)
		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Application cannot be empty")
	})
}

func TestResolver_EventingConfiguration(t *testing.T) {
	// GIVEN
	tnt := "tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	applicationID := uuid.New()
	gqlApp := fixGQLApplication(applicationID.String(), "bar", "baz")

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	defaultEveningURL := "https://eventing.domain.local"
	modelAppEventingCfg := fixModelApplicationEventingConfiguration(defaultEveningURL)
	gqlAppEventingCfg := fixGQLApplicationEventingConfiguration(defaultEveningURL)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn   func() *automock.EventingService
		ExpectedOutput  *graphql.ApplicationEventingConfiguration
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForApplication", txtest.CtxWithDBMatcher(), applicationID).Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: gqlAppEventingCfg,
			ExpectedError:  nil,
		}, {
			Name:            "Error when getting the configuration for runtime failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForApplication", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()

				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForApplication", txtest.CtxWithDBMatcher(), applicationID).Return(modelAppEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()

			resolver := application.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, eventingSvc)

			// WHEN
			result, err := resolver.EventingConfiguration(ctx, gqlApp)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, eventingSvc, transact, persist)
		})
	}

	t.Run("Error when parent object ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.EventingConfiguration(ctx, &graphql.Application{ID: "abc"})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing application ID as UUID")
		assert.Nil(t, result)
	})

	t.Run("Error when parent object is nil", func(t *testing.T) {
		// GIVEN
		resolver := application.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.EventingConfiguration(context.TODO(), nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Application cannot be empty")
		assert.Nil(t, result)
	})
}

func fixOAuths() []model.SystemAuth {
	return []model.SystemAuth{
		{
			ID:       "foo",
			TenantID: "foo",
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: nil,
					Oauth: &model.OAuthCredentialData{
						ClientID:     "foo",
						ClientSecret: "foo",
						URL:          "foo",
					},
				},
			},
		},
		{
			ID:       "bar",
			TenantID: "bar",
			Value:    nil,
		},
		{
			ID:       "test",
			TenantID: "test",
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "test",
						Password: "test",
					},
					Oauth: nil,
				},
			},
		},
	}
}
