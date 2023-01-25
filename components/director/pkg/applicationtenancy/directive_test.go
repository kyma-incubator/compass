package applicationtenancy_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/applicationtenancy"
	"github.com/kyma-incubator/compass/components/director/pkg/applicationtenancy/automock"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tenantID       = "381060d3-39ed-4617-a14a-f6fcf52a1e7e"
	newTenantID    = "12345678-1111-4617-0000-f6fcf52a1e7e"
	applicationID  = "637060ad-f30e-4326-a8f0-6dfae63d8dc9"
	parentTenantID = "f2cad1bc-caae-4d5b-8699-649c155a9939"
)

var (
	testErr                         = errors.New("test-err")
	subaccountTenantModel           = fixBusinessTenantMappingModel(tenantID, tenantpkg.Subaccount, false)
	rgTenantModel                   = fixBusinessTenantMappingModel(tenantID, tenantpkg.ResourceGroup, false)
	accountTenantModel              = fixBusinessTenantMappingModel(tenantID, tenantpkg.Account, true)
	newAccountTenantModel           = fixBusinessTenantMappingModel(newTenantID, tenantpkg.Account, true)
	accountTenantWithoutParentModel = fixBusinessTenantMappingModel(tenantID, tenantpkg.Account, false)
	customerTenantModel             = fixBusinessTenantMappingModel(parentTenantID, tenantpkg.Customer, false)
	orgTenantsModel                 = []*model.BusinessTenantMapping{fixBusinessTenantMappingModel(tenantID, tenantpkg.Organization, true)}
	applicationsModel               = []*model.Application{fixApplicationModel(applicationID)}
)

func TestDirective_TestSynchronizeApplicationTenancy(t *testing.T) {
	testCases := []struct {
		Name             string
		TxFn             func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn      func() *automock.BusinessTenantMappingService
		ApplicationSvcFn func() *automock.ApplicationService
		GetCtx           func() context.Context
		Resolver         func() func(ctx context.Context) (res interface{}, err error)
		EventType        schema.EventType
		ExpectedError    error
		ExpectedResult   interface{}
	}{
		{
			Name: "NEW_APPLICATION flow: Success",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(rgTenantModel, nil)
				tenantService.On("GetCustomerIDParentRecursively", txtest.CtxWithDBMatcher(), tenantID).Return(parentTenantID, nil)
				tenantService.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenantID).Return(customerTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Account).Return([]*model.BusinessTenantMapping{accountTenantModel}, nil)
				tenantService.On("CreateTenantAccessForResource", txtest.CtxWithDBMatcher(), tenantID, applicationID, false, resource.Application).Return(nil)

				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessGraphQLApplicationResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedResult:   mockedGraphQLApplicationNextOutput(),
		},
		{
			Name:             "Should not do anything when resolver fails",
			TxFn:             txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction,
			TenantSvcFn:      fixEmptyTenantService,
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixErrorResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    testErr,
		},
		{
			Name:             "Should fail when transaction fails to begin",
			TxFn:             txtest.NewTransactionContextGenerator(testErr).ThatFailsOnBegin,
			TenantSvcFn:      fixEmptyTenantService,
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    testErr,
		},
		{
			Name:        "NEW_APPLICATION flow: Should fail when there is no tenant in the context",
			TxFn:        txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: fixEmptyTenantService,
			GetCtx: func() context.Context {
				return context.TODO()
			},
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    errors.New("cannot read tenant from context"),
		},
		{
			Name: "NEW_APPLICATION flow: Should fail when getting tenant by ID",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_APPLICATION flow: Should not do anything when tenant type is not expected",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(subaccountTenantModel, nil)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedResult:   mockedTenantNextOutput(),
		},
		{
			Name: "NEW_APPLICATION flow: Should fail when parsing response to graphQL entity",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(rgTenantModel, nil)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    errors.New("Invalid data [reason=An error occurred while casting the response entity: 12345678-1111-4617-0000-f6fcf52a1e7e]"),
		},
		{
			Name: "NEW_APPLICATION flow: Should fail when getting parent customer ID",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(rgTenantModel, nil)
				tenantService.On("GetCustomerIDParentRecursively", txtest.CtxWithDBMatcher(), tenantID).Return("", testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessGraphQLApplicationResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_APPLICATION flow: Should fail when getting tenant by external ID",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(rgTenantModel, nil)
				tenantService.On("GetCustomerIDParentRecursively", txtest.CtxWithDBMatcher(), tenantID).Return(parentTenantID, nil)
				tenantService.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenantID).Return(nil, testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessGraphQLApplicationResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_APPLICATION flow: Should fail when listing tenant by parent and type",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(rgTenantModel, nil)
				tenantService.On("GetCustomerIDParentRecursively", txtest.CtxWithDBMatcher(), tenantID).Return(parentTenantID, nil)
				tenantService.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenantID).Return(customerTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Account).Return(nil, testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessGraphQLApplicationResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_APPLICATION flow: Should fail when creating tenant access for Application resources",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(rgTenantModel, nil)
				tenantService.On("GetCustomerIDParentRecursively", txtest.CtxWithDBMatcher(), tenantID).Return(parentTenantID, nil)
				tenantService.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenantID).Return(customerTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Account).Return([]*model.BusinessTenantMapping{accountTenantModel}, nil)
				tenantService.On("CreateTenantAccessForResource", txtest.CtxWithDBMatcher(), tenantID, applicationID, false, resource.Application).Return(testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessGraphQLApplicationResolver,
			EventType:        schema.EventTypeNewApplication,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_SINGLE_TENANT flow: Success",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(newAccountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(orgTenantsModel, nil)
				tenantService.On("CreateTenantAccessForResource", txtest.CtxWithDBMatcher(), newTenantID, applicationID, false, resource.Application).Return(nil)
				return tenantService
			},
			GetCtx: fixContextWithTenant,
			ApplicationSvcFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListAll", txtest.CtxWithDBMatcher()).Return(applicationsModel, nil)
				return appService
			},
			Resolver:       fixSuccessTenantResolver,
			EventType:      schema.EventTypeNewSingleTenant,
			ExpectedResult: mockedTenantNextOutput(),
		},
		{
			Name:             "NEW_SINGLE_TENANT flow: Should fail when parsing response to string",
			TxFn:             txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn:      fixEmptyTenantService,
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessGraphQLApplicationResolver,
			EventType:        schema.EventTypeNewSingleTenant,
			ExpectedError:    errors.New("An error occurred while casting the response entity"),
		},
		{
			Name: "NEW_SINGLE_TENANT flow: Should fail when getting tenant by ID",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewSingleTenant,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_SINGLE_TENANT flow: Should not do anything when new tenant is not Account",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(subaccountTenantModel, nil)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewSingleTenant,
			ExpectedResult:   mockedTenantNextOutput(),
		},
		{
			Name: "NEW_SINGLE_TENANT flow: Should not do anything when new tenant is Account but does not have a parent",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(accountTenantWithoutParentModel, nil)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewSingleTenant,
			ExpectedResult:   mockedTenantNextOutput(),
		},
		{
			Name: "NEW_SINGLE_TENANT flow: Should fail when listing tenants by parent and type errors",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(accountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(nil, testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessTenantResolver,
			EventType:        schema.EventTypeNewSingleTenant,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_SINGLE_TENANT flow: Should fail when listing applications for org tenant errors",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(accountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(orgTenantsModel, nil)
				return tenantService
			},
			GetCtx: fixContextWithTenant,
			ApplicationSvcFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListAll", txtest.CtxWithDBMatcher()).Return(nil, testErr)
				return appService
			},
			Resolver:      fixSuccessTenantResolver,
			EventType:     schema.EventTypeNewSingleTenant,
			ExpectedError: testErr,
		},
		{
			Name: "NEW_SINGLE_TENANT flow: Should fail when creating tenant access for application errors",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(newAccountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(orgTenantsModel, nil)
				tenantService.On("CreateTenantAccessForResource", txtest.CtxWithDBMatcher(), newTenantID, applicationID, false, resource.Application).Return(testErr)
				return tenantService
			},
			GetCtx: fixContextWithTenant,
			ApplicationSvcFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListAll", txtest.CtxWithDBMatcher()).Return(applicationsModel, nil)
				return appService
			},
			Resolver:      fixSuccessTenantResolver,
			EventType:     schema.EventTypeNewSingleTenant,
			ExpectedError: testErr,
		},
		{
			Name: "NEW_MULTIPLE_TENANTS flow: Success",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(newAccountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(orgTenantsModel, nil)
				tenantService.On("CreateTenantAccessForResource", txtest.CtxWithDBMatcher(), newTenantID, applicationID, false, resource.Application).Return(nil)
				return tenantService
			},
			GetCtx: fixContextWithTenant,
			ApplicationSvcFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListAll", txtest.CtxWithDBMatcher()).Return(applicationsModel, nil)
				return appService
			},
			Resolver:       fixSuccessMultipleTenantResolver,
			EventType:      schema.EventTypeNewMultipleTenants,
			ExpectedResult: mockedMultipleTenantsNextOutput(),
		},
		{
			Name:             "NEW_MULTIPLE_TENANTS flow: Should fail when parsing response to string array",
			TxFn:             txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn:      fixEmptyTenantService,
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessGraphQLApplicationResolver,
			EventType:        schema.EventTypeNewMultipleTenants,
			ExpectedError:    errors.New("An error occurred while casting the response entity"),
		},
		{
			Name: "NEW_MULTIPLE_TENANTS flow: Should fail getting tenant by ID errors",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessMultipleTenantResolver,
			EventType:        schema.EventTypeNewMultipleTenants,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_MULTIPLE_TENANTS flow: Should not do anything when tenant type is not Account",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(rgTenantModel, nil)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessMultipleTenantResolver,
			EventType:        schema.EventTypeNewMultipleTenants,
			ExpectedResult:   mockedMultipleTenantsNextOutput(),
		},
		{
			Name: "NEW_MULTIPLE_TENANTS flow: Should not do anything when tenant does not have parent",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(accountTenantWithoutParentModel, nil)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessMultipleTenantResolver,
			EventType:        schema.EventTypeNewMultipleTenants,
			ExpectedResult:   mockedMultipleTenantsNextOutput(),
		},
		{
			Name: "NEW_MULTIPLE_TENANTS flow: Should fail when listing tenants by parent and type errors",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(accountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(nil, testErr)
				return tenantService
			},
			GetCtx:           fixContextWithTenant,
			ApplicationSvcFn: fixEmptyApplicationService,
			Resolver:         fixSuccessMultipleTenantResolver,
			EventType:        schema.EventTypeNewMultipleTenants,
			ExpectedError:    testErr,
		},
		{
			Name: "NEW_MULTIPLE_TENANTS flow: Should fail when listing applications for org tenant errors",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(accountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(orgTenantsModel, nil)
				return tenantService
			},
			GetCtx: fixContextWithTenant,
			ApplicationSvcFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListAll", txtest.CtxWithDBMatcher()).Return(nil, testErr)
				return appService
			},
			Resolver:      fixSuccessMultipleTenantResolver,
			EventType:     schema.EventTypeNewMultipleTenants,
			ExpectedError: testErr,
		},
		{
			Name: "NEW_MULTIPLE_TENANTS flow: Should fail when creating tenant access for application errors",
			TxFn: txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(newAccountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(orgTenantsModel, nil)
				tenantService.On("CreateTenantAccessForResource", txtest.CtxWithDBMatcher(), newTenantID, applicationID, false, resource.Application).Return(testErr)
				return tenantService
			},
			GetCtx: fixContextWithTenant,
			ApplicationSvcFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListAll", txtest.CtxWithDBMatcher()).Return(applicationsModel, nil)
				return appService
			},
			Resolver:      fixSuccessMultipleTenantResolver,
			EventType:     schema.EventTypeNewMultipleTenants,
			ExpectedError: testErr,
		},
		{
			Name: "Should fail when committing transaction",
			TxFn: txtest.NewTransactionContextGenerator(testErr).ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantService := &automock.BusinessTenantMappingService{}
				tenantService.On("GetTenantByID", txtest.CtxWithDBMatcher(), newTenantID).Return(newAccountTenantModel, nil)
				tenantService.On("ListByParentAndType", txtest.CtxWithDBMatcher(), parentTenantID, tenantpkg.Organization).Return(orgTenantsModel, nil)
				tenantService.On("CreateTenantAccessForResource", txtest.CtxWithDBMatcher(), newTenantID, applicationID, false, resource.Application).Return(nil)
				return tenantService
			},
			GetCtx: fixContextWithTenant,
			ApplicationSvcFn: func() *automock.ApplicationService {
				appService := &automock.ApplicationService{}
				appService.On("ListAll", txtest.CtxWithDBMatcher()).Return(applicationsModel, nil)
				return appService
			},
			Resolver:      fixSuccessMultipleTenantResolver,
			EventType:     schema.EventTypeNewMultipleTenants,
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			appSvc := testCase.ApplicationSvcFn()
			persist, transact := testCase.TxFn()

			directive := applicationtenancy.NewDirective(transact, tenantSvc, appSvc)

			// WHEN
			res, err := directive.SynchronizeApplicationTenancy(testCase.GetCtx(), nil, testCase.Resolver(), testCase.EventType)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, res)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, appSvc)
		})
	}
}

type dummyResolver struct {
	called bool
}

func (d *dummyResolver) SuccessTenantResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return mockedTenantNextOutput(), nil
}

func (d *dummyResolver) SuccessMultipleTenantResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return mockedMultipleTenantsNextOutput(), nil
}

func (d *dummyResolver) ErrorResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return nil, testErr
}

func (d *dummyResolver) SuccessGraphQLApplicationResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return mockedGraphQLApplicationNextOutput(), nil
}

func mockedGraphQLApplicationNextOutput() *schema.Application {
	return &schema.Application{
		BaseEntity: &schema.BaseEntity{
			ID:    applicationID,
			Ready: true,
		},
	}
}

func mockedTenantNextOutput() string {
	return newTenantID
}

func mockedMultipleTenantsNextOutput() []string {
	return []string{newTenantID}
}

func fixSuccessGraphQLApplicationResolver() func(_ context.Context) (res interface{}, err error) {
	r := &dummyResolver{}
	return r.SuccessGraphQLApplicationResolve
}

func fixSuccessTenantResolver() func(_ context.Context) (res interface{}, err error) {
	r := &dummyResolver{}
	return r.SuccessTenantResolve
}

func fixSuccessMultipleTenantResolver() func(_ context.Context) (res interface{}, err error) {
	r := &dummyResolver{}
	return r.SuccessMultipleTenantResolve
}

func fixErrorResolver() func(_ context.Context) (res interface{}, err error) {
	r := &dummyResolver{}
	return r.ErrorResolve
}

func fixBusinessTenantMappingModel(id string, tntType tenantpkg.Type, hasParent bool) *model.BusinessTenantMapping {
	parent := ""
	if hasParent {
		parent = parentTenantID
	}
	return &model.BusinessTenantMapping{
		ID:             id,
		ExternalTenant: id,
		Type:           tntType,
		Parent:         parent,
	}
}

func fixApplicationModel(id string) *model.Application {
	return &model.Application{
		Name: "test-app",
		BaseEntity: &model.BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}

func fixEmptyApplicationService() *automock.ApplicationService {
	return &automock.ApplicationService{}
}

func fixEmptyTenantService() *automock.BusinessTenantMappingService {
	return &automock.BusinessTenantMappingService{}
}

func fixContextWithTenant() context.Context {
	return context.WithValue(context.TODO(), tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
}
