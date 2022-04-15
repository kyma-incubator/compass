package statusupdate_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	logrustest "github.com/sirupsen/logrus/hooks/test"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/statusupdate"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/statusupdate/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
)

func TestUpdate_Handler(t *testing.T) {
	// GIVEN
	testErr := errors.New("test")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                 string
		TxFn                 func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RepoFn               func() *automock.StatusUpdateRepository
		Request              *http.Request
		ExpectedStatus       int
		ExpectedResponse     string
		ExpectedErrorMessage *string
		ExpectedError        *string
		MockNextHandler      http.Handler
	}{
		{
			Name: "In case of Integration System do nothing and execute next handler",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(testID, consumer.IntegrationSystem, oathkeeper.OAuth2Flow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "In case of Static User do nothing and execute next handler",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(testID, consumer.User, oathkeeper.JWTAuthFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "In case of Application and Certificate flow update status and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(false, nil)
				repo.On("UpdateStatus", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(nil)
				return &repo
			},
			Request:          createRequestWithClaims(testID, consumer.Application, oathkeeper.CertificateFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "In case of Application and OneTimeToken flow do nothing and execute next handler",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(testID, consumer.Application, oathkeeper.OneTimeTokenFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "In case of Runtime and Certificate flow update status and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Runtimes).Return(false, nil)
				repo.On("UpdateStatus", txtest.CtxWithDBMatcher(), testID, statusupdate.Runtimes).Return(nil)
				return &repo
			},
			Request:          createRequestWithClaims(testID, consumer.Runtime, oathkeeper.CertificateFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "In case of Runtime and OneTimeToken flow do nothing and execute next handler",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(testID, consumer.Runtime, oathkeeper.OneTimeTokenFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "In case of application already connected do nothing and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, statusupdate.Applications).Return(true, nil)
				return &repo
			},
			Request:          createRequestWithClaims(testID, consumer.Application, oathkeeper.CertificateFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "In case of runtime already connected do nothing and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Runtimes).Return(true, nil)
				return &repo
			},
			Request:          createRequestWithClaims(testID, consumer.Runtime, oathkeeper.CertificateFlow),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t),
		},
		{
			Name: "Error when no consumer info in context",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				return &repo
			},
			Request:              &http.Request{},
			ExpectedStatus:       http.StatusOK,
			ExpectedErrorMessage: str.Ptr("An error has occurred while fetching consumer info from context:"),
			ExpectedError:        str.Ptr("Internal Server Error: cannot read consumer from context"),
			MockNextHandler:      fixNextHandler(t),
		},
		{
			Name: "Error when starting transaction",
			TxFn: txGen.ThatFailsOnBegin,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				return &repo
			},
			Request:              createRequestWithClaims(testID, consumer.Application, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusOK,
			ExpectedErrorMessage: str.Ptr("An error has occurred while opening transaction:"),
			ExpectedError:        str.Ptr("test"),
			MockNextHandler:      fixNextHandler(t),
		},
		{
			Name: "Error when checking if already connected",
			TxFn: txGen.ThatDoesntExpectCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(false, testErr)
				return &repo
			},
			Request:              createRequestWithClaims(testID, consumer.Application, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusOK,
			ExpectedErrorMessage: str.Ptr("An error has occurred while checking repository status:"),
			ExpectedError:        str.Ptr("test"),
			MockNextHandler:      fixNextHandler(t),
		},
		{
			Name: "Error when failing on commit",
			TxFn: txGen.ThatFailsOnCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(true, nil)
				return &repo
			},
			Request:              createRequestWithClaims(testID, consumer.Application, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusOK,
			ExpectedErrorMessage: str.Ptr("An error has occurred while committing transaction:"),
			ExpectedError:        str.Ptr("test"),
			MockNextHandler:      fixNextHandler(t),
		},
		{
			Name: "Error when updating status",
			TxFn: txGen.ThatDoesntExpectCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(false, nil)
				repo.On("UpdateStatus", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(testErr)
				return &repo
			},
			Request:              createRequestWithClaims(testID, consumer.Application, oathkeeper.CertificateFlow),
			ExpectedStatus:       http.StatusOK,
			ExpectedErrorMessage: str.Ptr("An error has occurred while updating repository status:"),
			ExpectedError:        str.Ptr("test"),
			MockNextHandler:      fixNextHandler(t),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepoFn()
			persist, transact := testCase.TxFn()
			var actualLog bytes.Buffer
			logger, hook := logrustest.NewNullLogger()
			logger.SetFormatter(&logrus.TextFormatter{
				DisableTimestamp: true,
			})
			logger.SetOutput(&actualLog)
			ctx := log.ContextWithLogger(testCase.Request.Context(), logrus.NewEntry(logger))
			update := statusupdate.New(transact, repo)
			req := testCase.Request.WithContext(ctx)
			// WHEN
			rr := httptest.NewRecorder()
			updateHandler := update.Handler()
			updateHandler(testCase.MockNextHandler).ServeHTTP(rr, req)

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
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}
}

func createRequestWithClaims(id string, consumerType consumer.ConsumerType, flow oathkeeper.AuthFlow) *http.Request {
	req := http.Request{}
	apiConsumer := consumer.Consumer{ConsumerID: id, ConsumerType: consumerType, Flow: flow}
	ctxWithConsumerInfo := consumer.SaveToContext(context.TODO(), apiConsumer)
	return req.WithContext(ctxWithConsumerInfo)
}

func fixNextHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}
