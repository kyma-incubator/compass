package appmetadatavalidation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/appmetadatavalidation"
	"github.com/kyma-incubator/compass/components/director/internal/appmetadatavalidation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
}

type dummyResolver struct {
	called bool
}

func (d *dummyResolver) SuccessResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return mockedNextOutput(), nil
}

func mockedNextOutput() string {
	return "nextOutput"
}

func TestDirective_Handler(t *testing.T) {
	// GIVEN
	testErr := errors.New("test")
	testErrRegionMismatch := errors.New("labels mismatch: \"eu-1\" and \"fake\"")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	consumerID := "0191fcfd-ae7e-4d1a-8027-520a96d5319f"
	consumerIDInternal := "11111111-ae7e-4d1a-8027-520a96d5319f"
	regionLabelValue := "eu-1"
	tenantHeaderID := "abcd1122-ae7e-4d1a-8027-520a96d5319d"
	tenantHeaderIDInternal := "2222222-ae7e-4d1a-8027-520a96d5319f"

	consumerModel := fixBusinessTenantMappingModel(consumerIDInternal, consumerID)
	tenantHeaderModel := fixBusinessTenantMappingModel(tenantHeaderIDInternal, tenantHeaderID)
	consumerLabel := fixTenantLabel(consumerIDInternal, regionLabelValue)
	tenantHeaderLabel := fixTenantLabel(tenantHeaderIDInternal, regionLabelValue)
	tenantHeaderLabelWithDifferentRegion := fixTenantLabel(tenantHeaderIDInternal, "fake")

	ts := testStruct{}

	testCases := []struct {
		Name          string
		TxFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn   func() *automock.TenantService
		LabelSvcFn    func() *automock.LabelService
		Context       context.Context
		ExpectedError *string
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.On("GetTenantByExternalID", mock.Anything, consumerID).Return(consumerModel, nil).Once()
				tntSvc.On("GetTenantByExternalID", mock.Anything, tenantHeaderID).Return(tenantHeaderModel, nil).Once()
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, consumerIDInternal, model.TenantLabelableObject, consumerIDInternal, tenant.RegionLabelKey).Return(consumerLabel, nil).Once()
				lblSvc.On("GetByKey", mock.Anything, tenantHeaderIDInternal, model.TenantLabelableObject, tenantHeaderIDInternal, tenant.RegionLabelKey).Return(tenantHeaderLabel, nil).Once()
				return lblSvc
			},
			Context: ctxWithTenant(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
		},
		{
			Name: "Success when no consumer is provided",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.AssertNotCalled(t, "GetTenantByExternalID")
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.AssertNotCalled(t, "GetByKey")
				return lblSvc
			},
			Context: context.TODO(),
		},
		{
			Name: "Success when the flow is not cert",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.AssertNotCalled(t, "GetTenantByExternalID")
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.AssertNotCalled(t, "GetByKey")
				return lblSvc
			},
			Context: ctxWithTenant(tenantHeaderID, consumerID, consumer.Application, oathkeeper.OneTimeTokenFlow),
		},
		{
			Name: "Success when flow is cert but tenant header is missing",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.AssertNotCalled(t, "GetTenantByExternalID")
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.AssertNotCalled(t, "GetByKey")
				return lblSvc
			},
			Context: ctxWithTenantAndEmptyHeader(consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
		},
		{
			Name: "Error when starting transaction",
			TxFn: txGen.ThatFailsOnBegin,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.AssertNotCalled(t, "GetTenantByExternalID")
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.AssertNotCalled(t, "GetByKey")
				return lblSvc
			},
			Context:       ctxWithTenant(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedError: str.Ptr(testErr.Error()),
		},
		{
			Name: "Error when fetching consumer tenant",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.On("GetTenantByExternalID", mock.Anything, consumerID).Return(nil, testErr).Once()
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.AssertNotCalled(t, "GetByKey")
				return lblSvc
			},
			Context:       ctxWithTenant(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedError: str.Ptr(testErr.Error()),
		},
		{
			Name: "Error when fetching consumer tenant region label",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.On("GetTenantByExternalID", mock.Anything, consumerID).Return(consumerModel, nil).Once()
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, consumerIDInternal, model.TenantLabelableObject, consumerIDInternal, tenant.RegionLabelKey).Return(nil, testErr).Once()

				return lblSvc
			},
			Context:       ctxWithTenant(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedError: str.Ptr(testErr.Error()),
		},
		{
			Name: "Error when fetching header tenant",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.On("GetTenantByExternalID", mock.Anything, consumerID).Return(consumerModel, nil).Once()
				tntSvc.On("GetTenantByExternalID", mock.Anything, tenantHeaderID).Return(nil, testErr).Once()
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, consumerIDInternal, model.TenantLabelableObject, consumerIDInternal, tenant.RegionLabelKey).Return(consumerLabel, nil).Once()

				return lblSvc
			},
			Context:       ctxWithTenant(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedError: str.Ptr(testErr.Error()),
		},
		{
			Name: "Error when fetching header tenant region label",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.On("GetTenantByExternalID", mock.Anything, consumerID).Return(consumerModel, nil).Once()
				tntSvc.On("GetTenantByExternalID", mock.Anything, tenantHeaderID).Return(tenantHeaderModel, nil).Once()
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, consumerIDInternal, model.TenantLabelableObject, consumerIDInternal, tenant.RegionLabelKey).Return(consumerLabel, nil).Once()
				lblSvc.On("GetByKey", mock.Anything, tenantHeaderIDInternal, model.TenantLabelableObject, tenantHeaderIDInternal, tenant.RegionLabelKey).Return(nil, testErr).Once()

				return lblSvc
			},
			Context:       ctxWithTenant(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedError: str.Ptr(testErr.Error()),
		},
		{
			Name: "Error when regions do not match",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.TenantService {
				tntSvc := &automock.TenantService{}
				tntSvc.On("GetTenantByExternalID", mock.Anything, consumerID).Return(consumerModel, nil).Once()
				tntSvc.On("GetTenantByExternalID", mock.Anything, tenantHeaderID).Return(tenantHeaderModel, nil).Once()
				return tntSvc
			},
			LabelSvcFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, consumerIDInternal, model.TenantLabelableObject, consumerIDInternal, tenant.RegionLabelKey).Return(consumerLabel, nil).Once()
				lblSvc.On("GetByKey", mock.Anything, tenantHeaderIDInternal, model.TenantLabelableObject, tenantHeaderIDInternal, tenant.RegionLabelKey).Return(tenantHeaderLabelWithDifferentRegion, nil).Once()

				return lblSvc
			},
			Context:       ctxWithTenant(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedError: str.Ptr(testErrRegionMismatch.Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			labelSvc := testCase.LabelSvcFn()
			persist, transact := testCase.TxFn()

			dummyResolver := dummyResolver{}

			handler := appmetadatavalidation.NewDirective(transact, tenantSvc, labelSvc)
			// WHEN
			res, err := handler.Validate(testCase.Context, ts, dummyResolver.SuccessResolve)
			// THEN

			if testCase.ExpectedError != nil {
				require.NotEmpty(t, err)
				require.Contains(t, err.Error(), *testCase.ExpectedError)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, res, mockedNextOutput())
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, labelSvc)
		})
	}
}

func ctxWithTenant(tenantID, consumerID string, consumerType consumer.ConsumerType, flow oathkeeper.AuthFlow) context.Context {
	apiConsumer := consumer.Consumer{ConsumerID: consumerID, ConsumerType: consumerType, Flow: flow}
	ctx := context.TODO()
	if len(tenantID) > 0 {
		ctx = context.WithValue(ctx, appmetadatavalidation.TenantHeader, tenantID)
	}
	return consumer.SaveToContext(ctx, apiConsumer)
}

func ctxWithTenantAndEmptyHeader(consumerID string, consumerType consumer.ConsumerType, flow oathkeeper.AuthFlow) context.Context {
	apiConsumer := consumer.Consumer{ConsumerID: consumerID, ConsumerType: consumerType, Flow: flow}
	ctx := context.TODO()
	context.WithValue(ctx, appmetadatavalidation.TenantHeader, "")
	return consumer.SaveToContext(ctx, apiConsumer)
}
