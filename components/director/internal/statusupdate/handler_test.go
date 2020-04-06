package statusupdate_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/statusupdate"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/statusupdate/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
)

const (
	applicationsTable = "applications"
	runtimesTable     = "runtimes"
)

func TestUpdate_Handler(t *testing.T) {
	//given
	testErr := errors.New("test")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TxFn             func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RepoFn           func() *automock.StatusUpdateRepository
		Request          *http.Request
		ExpectedStatus   int
		ExpectedResponse string
		ExpectedLog      bytes.Buffer
		MockNextHandler  http.Handler
	}{
		{
			Name: "In case of Integration System do nothing and execute next handler",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(t, testID, consumer.IntegrationSystem),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t, testID, consumer.IntegrationSystem),
		},
		{
			Name: "In case of Static User do nothing and execute next handler",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(t, testID, consumer.User),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t, testID, consumer.User),
		},
		{
			Name: "In case of Application update status and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(false, nil)
				repo.On("UpdateStatus", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t, testID, consumer.Application),
		},
		{
			Name: "In case of Runtime update status and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Runtimes).Return(false, nil)
				repo.On("UpdateStatus", txtest.CtxWithDBMatcher(), testID, statusupdate.Runtimes).Return(nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Runtime),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t, testID, consumer.Runtime),
		},
		{
			Name: "In case of application already connected do nothing and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, statusupdate.Applications).Return(true, nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t, testID, consumer.Application),
		},
		{
			Name: "In case of runtime already connected do nothing and execute next handler",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Runtimes).Return(true, nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Runtime),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockNextHandler:  fixNextHandler(t, testID, consumer.Runtime),
		},
		{
			Name: "Error when no consumer info in context",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				return &repo
			},
			Request:         &http.Request{},
			ExpectedStatus:  http.StatusOK,
			ExpectedLog:     *bytes.NewBufferString("while fetching consumer info from from context: cannot read consumer from context"),
			MockNextHandler: fixNextHandler(t, testID, consumer.IntegrationSystem),
		},
		{
			Name: "Error when starting transaction",
			TxFn: txGen.ThatFailsOnBegin,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				return &repo
			},
			Request:         createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:  http.StatusOK,
			ExpectedLog:     *bytes.NewBufferString("while opening transaction: test"),
			MockNextHandler: fixNextHandler(t, testID, consumer.Application),
		},
		{
			Name: "Error when checking if already connected",
			TxFn: txGen.ThatDoesntExpectCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(false, testErr)
				return &repo
			},
			Request:         createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:  http.StatusOK,
			ExpectedLog:     *bytes.NewBufferString("while checking status: test"),
			MockNextHandler: fixNextHandler(t, testID, consumer.Application),
		},
		{
			Name: "Error when failing on commit",
			TxFn: txGen.ThatFailsOnCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", txtest.CtxWithDBMatcher(), testID, statusupdate.Applications).Return(true, nil)
				return &repo
			},
			Request:         createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:  http.StatusOK,
			ExpectedLog:     *bytes.NewBufferString("while committing: test"),
			MockNextHandler: fixNextHandler(t, testID, consumer.Application),
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
			Request:         createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:  http.StatusOK,
			ExpectedLog:     *bytes.NewBufferString("while updating status: test"),
			MockNextHandler: fixNextHandler(t, testID, consumer.Application),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepoFn()
			persist, transact := testCase.TxFn()
			var actualLog bytes.Buffer
			logger := logrus.New()
			logger.SetFormatter(&logrus.TextFormatter{
				DisableTimestamp: true,
			})
			logger.SetOutput(&actualLog)
			update := statusupdate.New(transact, repo, logger)

			// WHEN
			rr := httptest.NewRecorder()
			updateHandler := update.Handler()
			updateHandler(testCase.MockNextHandler).ServeHTTP(rr, testCase.Request)

			// THEN
			response := rr.Body.String()
			assert.Equal(t, testCase.ExpectedStatus, rr.Code)
			if testCase.ExpectedResponse == "OK" {
				assert.Equal(t, testCase.ExpectedResponse, response)
			}
			if testCase.ExpectedLog.String() != "" {
				expectedLog := fmt.Sprintf("level=error msg=\"%s\"\n", testCase.ExpectedLog.String())
				assert.Equal(t, expectedLog, actualLog.String())
			}
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}

}

func createRequestWithClaims(t *testing.T, id string, consumerType consumer.ConsumerType) *http.Request {
	req := http.Request{}
	apiConsumer := consumer.Consumer{ConsumerID: id, ConsumerType: consumerType}
	ctxWithConsumerInfo := consumer.SaveToContext(context.TODO(), apiConsumer)
	return req.WithContext(ctxWithConsumerInfo)
}

func fixNextHandler(t *testing.T, consumerID string, consumerType consumer.ConsumerType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}
