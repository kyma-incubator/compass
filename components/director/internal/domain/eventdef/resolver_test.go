package eventdef_test

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

var contextParam = txtest.CtxWithDBMatcher()

func TestResolver_AddEventAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	appId := "1"

	modelAPI := fixMinModelEventAPIDefinition(id, "placeholder")
	gqlAPI := fixGQLEventDefinition(id, "placeholder")
	gqlAPIInput := fixGQLEventDefinitionInput()
	modelAPIInput := fixModelEventDefinitionInput()

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn        func() *automock.EventDefService
		AppServiceFn     func() *automock.ApplicationService
		ConverterFn      func() *automock.EventAPIConverter
		ExpectedEventDef *graphql.EventDefinition
		ExpectedErr      error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Create", contextParam, appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPI, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			ExpectedEventDef: gqlAPI,
			ExpectedErr:      nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when application not exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      errors.New("Cannot add Event Definition to not existing Application"),
		},
		{
			Name:            "Returns error when application existence check failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(false, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when Event Definition creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Create", contextParam, appId, *modelAPIInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when Event Definition retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Create", contextParam, appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Create", contextParam, appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPI, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persistance, tx := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			appSvc := testCase.AppServiceFn()

			resolver := eventdef.NewResolver(tx, svc, appSvc, nil, converter, nil)

			// when
			result, err := resolver.AddEventDefinition(context.TODO(), appId, *gqlAPIInput)

			// then
			assert.Equal(t, testCase.ExpectedEventDef, result)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persistance.AssertExpectations(t)
			tx.AssertExpectations(t)
			svc.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddEventAPIToPackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	packageID := "1"

	modelAPI := fixMinModelEventAPIDefinition(id, "placeholder")
	gqlAPI := fixGQLEventDefinition(id, "placeholder")
	gqlAPIInput := fixGQLEventDefinitionInput()
	modelAPIInput := fixModelEventDefinitionInput()

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn        func() *automock.EventDefService
		PkgServiceFn     func() *automock.PackageService
		ConverterFn      func() *automock.EventAPIConverter
		ExpectedEventDef *graphql.EventDefinition
		ExpectedErr      error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateToPackage", contextParam, packageID, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPI, nil).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			ExpectedEventDef: gqlAPI,
			ExpectedErr:      nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when application not exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      errors.New("Cannot add Event Definition to not existing Package"),
		},
		{
			Name:            "Returns error when application existence check failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(false, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when Event Definition creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateToPackage", contextParam, packageID, *modelAPIInput).Return("", testErr).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when Event Definition retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateToPackage", contextParam, packageID, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateToPackage", contextParam, packageID, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPI, nil).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persistance, tx := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			pkgSvc := testCase.PkgServiceFn()

			resolver := eventdef.NewResolver(tx, svc, nil, pkgSvc, converter, nil)

			// when
			result, err := resolver.AddEventDefinitionToPackage(context.TODO(), packageID, *gqlAPIInput)

			// then
			assert.Equal(t, testCase.ExpectedEventDef, result)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persistance.AssertExpectations(t)
			tx.AssertExpectations(t)
			svc.AssertExpectations(t)
			pkgSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteEventAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelAPIDefinition := fixMinModelEventAPIDefinition(id, "placeholder")
	gqlAPIDefinition := fixGQLEventDefinition(id, "placeholder")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn        func() *automock.EventDefService
		ConverterFn      func() *automock.EventAPIConverter
		ExpectedEventDef *graphql.EventDefinition
		ExpectedErr      error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", contextParam, id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", contextParam, id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedEventDef: gqlAPIDefinition,
			ExpectedErr:      nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when Event Definition retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when API deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", contextParam, id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", contextParam, id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", contextParam, id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", contextParam, id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persistance, tx := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := eventdef.NewResolver(tx, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.DeleteEventDefinition(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedEventDef, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persistance.AssertExpectations(t)
			tx.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateEventAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	gqlAPIDefinitionInput := fixGQLEventDefinitionInput()
	modelAPIDefinitionInput := fixModelEventDefinitionInput()
	gqlAPIDefinition := fixGQLEventDefinition(id, "placeholder")
	modelAPIDefinition := fixMinModelEventAPIDefinition(id, "placeholder")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn        func() *automock.EventDefService
		ConverterFn      func() *automock.EventAPIConverter
		InputWebhookID   string
		InputDefinition  graphql.EventDefinitionInput
		ExpectedEventDef *graphql.EventDefinition
		ExpectedErr      error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", contextParam, id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			InputWebhookID:   id,
			InputDefinition:  *gqlAPIDefinitionInput,
			ExpectedEventDef: gqlAPIDefinition,
			ExpectedErr:      nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when Event Definition update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", contextParam, id, *modelAPIDefinitionInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				return conv
			},
			InputWebhookID:   id,
			InputDefinition:  *gqlAPIDefinitionInput,
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when Event Definition retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", contextParam, id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				return conv
			},
			InputWebhookID:   id,
			InputDefinition:  *gqlAPIDefinitionInput,
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", contextParam, id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedEventDef: nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persistance, tx := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := eventdef.NewResolver(tx, svc, nil, nil, converter, nil)

			// when
			result, err := resolver.UpdateEventDefinition(context.TODO(), id, *gqlAPIDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedEventDef, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persistance.AssertExpectations(t)
			tx.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("test error")

	apiID := "apiID"

	dataBytes := "data"
	modelEventAPISpec := &model.EventSpec{
		Data: &dataBytes,
	}

	modelEventAPIDefinition := &model.EventDefinition{
		Spec: modelEventAPISpec,
	}

	clob := graphql.CLOB(dataBytes)
	gqlEventSpec := &graphql.EventSpec{
		Data: &clob,
	}

	gqlEventDefinition := &graphql.EventDefinition{
		Spec: gqlEventSpec,
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)
	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.EventDefService
		ConvFn            func() *automock.EventAPIConverter
		ExpectedEventSpec *graphql.EventSpec
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("RefetchAPISpec", txtest.CtxWithDBMatcher(), apiID).Return(modelEventAPISpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelEventAPIDefinition).Return(gqlEventDefinition).Once()
				return conv
			},
			ExpectedEventSpec: gqlEventSpec,
			ExpectedErr:       nil,
		},
		{
			Name:            "Retuns error when transaction commit faied",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("RefetchAPISpec", txtest.CtxWithDBMatcher(), apiID).Return(modelEventAPISpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
		{
			Name:            "Returns error when refetching Event spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("RefetchAPISpec", txtest.CtxWithDBMatcher(), apiID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
		{
			Name:            "Returns error when transaction start failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			ConvFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			conv := testCase.ConvFn()
			resolver := eventdef.NewResolver(transact, svc, nil, nil, conv, nil)

			// when
			result, err := resolver.RefetchEventDefinitionSpec(context.TODO(), apiID)

			// then
			assert.Equal(t, testCase.ExpectedEventSpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			conv.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}
}

func TestResolver_FetchRequest(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	url := "foo.bar"
	eventAPISpec := &graphql.EventSpec{DefinitionID: id}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	timestamp := time.Now()
	frModel := fixModelFetchRequest("foo", url, timestamp)
	frGQL := fixGQLFetchRequest(url, timestamp)
	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventDefService
		ConverterFn     func() *automock.FetchRequestConverter
		EventApiSpec    *graphql.EventSpec
		ExpectedResult  *graphql.FetchRequest
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("GetFetchRequest", contextParam, id).Return(frModel, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frModel).Return(frGQL).Once()
				return conv
			},
			EventApiSpec:   eventAPISpec,
			ExpectedResult: frGQL,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			EventApiSpec:   eventAPISpec,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Doesn't exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("GetFetchRequest", contextParam, id).Return(nil, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			EventApiSpec:   eventAPISpec,
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name:            "Error",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("GetFetchRequest", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			EventApiSpec:   eventAPISpec,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), id).Return(frModel, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			EventApiSpec:   eventAPISpec,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when obj is nil",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			EventApiSpec:   nil,
			ExpectedResult: nil,
			ExpectedErr:    errors.New("Event Spec cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := eventdef.NewResolver(transact, svc, nil, nil, nil, converter)

			// when
			result, err := resolver.FetchRequest(context.TODO(), testCase.EventApiSpec)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
