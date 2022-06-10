package appmetadatavalidation_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/appmetadatavalidation"
	"github.com/kyma-incubator/compass/components/director/internal/appmetadatavalidation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdate_Handler(t *testing.T) {
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

	testCases := []struct {
		Name                 string
		TxFn                 func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn          func() *automock.TenantService
		LabelSvcFn           func() *automock.LabelService
		Request              *http.Request
		ExpectedStatus       int
		ExpectedResponse     string
		ExpectedErrorMessage *string
		ExpectedError        *string
		MockNextHandler      http.Handler
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
			Request:          createRequestWithClaims(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
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
			Request:          createRequestWithNoClaims(),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
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
			Request:          createRequestWithClaims(tenantHeaderID, consumerID, consumer.Application, oathkeeper.OneTimeTokenFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
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
			Request:          createRequestWithClaims("", consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
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
			Request:              createRequestWithClaims(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusInternalServerError,
			ExpectedErrorMessage: str.Ptr("An error has occurred while opening transaction:"),
			ExpectedError:        str.Ptr(testErr.Error()),
			MockNextHandler:      fixNextHandler(t),
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
			Request:              createRequestWithClaims(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusInternalServerError,
			ExpectedErrorMessage: str.Ptr("An error has occurred while fetching tenant by external ID \"0191fcfd-ae7e-4d1a-8027-520a96d5319f\":"),
			ExpectedError:        str.Ptr(testErr.Error()),
			MockNextHandler:      fixNextHandler(t),
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
			Request:              createRequestWithClaims(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusInternalServerError,
			ExpectedErrorMessage: str.Ptr("An error has occurred while fetching \"region\" label for tenant ID \"0191fcfd-ae7e-4d1a-8027-520a96d5319f\":"),
			ExpectedError:        str.Ptr(testErr.Error()),
			MockNextHandler:      fixNextHandler(t),
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
			Request:              createRequestWithClaims(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusInternalServerError,
			ExpectedErrorMessage: str.Ptr("An error has occurred while fetching tenant by external ID \"abcd1122-ae7e-4d1a-8027-520a96d5319d\":"),
			ExpectedError:        str.Ptr(testErr.Error()),
			MockNextHandler:      fixNextHandler(t),
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
			Request:              createRequestWithClaims(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusInternalServerError,
			ExpectedErrorMessage: str.Ptr("An error has occurred while fetching \"region\" label for tenant ID \"abcd1122-ae7e-4d1a-8027-520a96d5319d\":"),
			ExpectedError:        str.Ptr(testErrRegionMismatch.Error()),
			MockNextHandler:      fixNextHandler(t),
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
			Request:         createRequestWithClaims(tenantHeaderID, consumerID, consumer.ExternalCertificate, oathkeeper.CertificateFlow),
			ExpectedStatus:  http.StatusInternalServerError,
			MockNextHandler: fixNextHandler(t),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			labelSvc := testCase.LabelSvcFn()
			persist, transact := testCase.TxFn()
			var actualLog bytes.Buffer
			logger, hook := logrustest.NewNullLogger()
			logger.SetFormatter(&logrus.TextFormatter{
				DisableTimestamp: true,
			})
			logger.SetOutput(&actualLog)
			ctx := log.ContextWithLogger(testCase.Request.Context(), logrus.NewEntry(logger))

			handler := appmetadatavalidation.NewHandler(transact, tenantSvc, labelSvc)
			req := testCase.Request.WithContext(ctx)
			// WHEN
			rr := httptest.NewRecorder()
			validationHandler := handler.Handler()
			validationHandler(testCase.MockNextHandler).ServeHTTP(rr, req)

			// THEN
			response := rr.Body.String()
			assert.Equal(t, testCase.ExpectedStatus, rr.Code)
			if testCase.ExpectedResponse == "OK" {
				assert.Equal(t, testCase.ExpectedResponse, response)
			}
			if testCase.ExpectedErrorMessage != nil {
				assert.Equal(t, *testCase.ExpectedErrorMessage+" "+*testCase.ExpectedError, hook.LastEntry().Message)
			}
			if testCase.ExpectedError != nil {
				assert.Equal(t, *testCase.ExpectedError, hook.LastEntry().Data["error"].(error).Error())
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, labelSvc)
		})
	}
}

func createRequestWithClaims(tenantHeaderID, consumerID string, consumerType consumer.ConsumerType, flow oathkeeper.AuthFlow) *http.Request {
	req := http.Request{}
	apiConsumer := consumer.Consumer{ConsumerID: consumerID, ConsumerType: consumerType, Flow: flow}
	ctxWithConsumerInfo := consumer.SaveToContext(context.TODO(), apiConsumer)
	req.Header = map[string][]string{}
	if len(tenantHeaderID) > 0 {
		req.Header.Set("tenant", tenantHeaderID)
	}
	return req.WithContext(ctxWithConsumerInfo)
}

func createRequestWithNoClaims() *http.Request {
	req := http.Request{}
	ctxWithConsumerInfo := context.TODO()
	req.Header = map[string][]string{}
	return req.WithContext(ctxWithConsumerInfo)
}

func fixNextHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}
