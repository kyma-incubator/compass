package statusupdate_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
		MockHandler      http.Handler
	}{
		{
			Name: "Success when integration system",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(t, testID, consumer.IntegrationSystem),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockHandler:      testHandler(t, testID, consumer.IntegrationSystem),
		},
		{
			Name: "Success when user",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				return nil
			},
			Request:          createRequestWithClaims(t, testID, consumer.User),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockHandler:      testHandler(t, testID, consumer.User),
		},
		{
			Name: "Success when application",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, applicationsTable).Return(false, nil)
				repo.On("UpdateStatus", mock.Anything, testID, applicationsTable).Return(nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockHandler:      testHandler(t, testID, consumer.Application),
		},
		{
			Name: "Success when runtime",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, runtimesTable).Return(false, nil)
				repo.On("UpdateStatus", mock.Anything, testID, runtimesTable).Return(nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Runtime),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockHandler:      testHandler(t, testID, consumer.Runtime),
		},
		{
			Name: "Success when application already connected",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, applicationsTable).Return(true, nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockHandler:      testHandler(t, testID, consumer.Application),
		},
		{
			Name: "Success when runtime already connected",
			TxFn: txGen.ThatSucceeds,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, runtimesTable).Return(true, nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Runtime),
			ExpectedStatus:   http.StatusOK,
			ExpectedResponse: "OK",
			MockHandler:      testHandler(t, testID, consumer.Runtime),
		},
		{
			Name: "Error when no consumer info in context",
			TxFn: txGen.ThatDoesntStartTransaction,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				return &repo
			},
			Request:          &http.Request{},
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedResponse: "{\"errors\":[{\"message\":\"while fetching consumer info from from context: cannot read consumer from context\"}]}\n",
			MockHandler:      testHandler(t, testID, consumer.IntegrationSystem),
		},
		{
			Name: "Error when starting transaction",
			TxFn: txGen.ThatFailsOnBegin,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: "{\"errors\":[{\"message\":\"while opening transaction: test\"}]}\n",
			MockHandler:      testHandler(t, testID, consumer.Application),
		},
		{
			Name: "Error when checking if already connected",
			TxFn: txGen.ThatDoesntExpectCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, applicationsTable).Return(false, testErr)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: "{\"errors\":[{\"message\":\"while checking status: test\"}]}\n",
			MockHandler:      testHandler(t, testID, consumer.Application),
		},
		{
			Name: "Error when failing on commit",
			TxFn: txGen.ThatFailsOnCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, applicationsTable).Return(true, nil)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: "{\"errors\":[{\"message\":\"while committing: test\"}]}\n",
			MockHandler:      testHandler(t, testID, consumer.Application),
		},
		{
			Name: "Error when updating",
			TxFn: txGen.ThatDoesntExpectCommit,
			RepoFn: func() *automock.StatusUpdateRepository {
				repo := automock.StatusUpdateRepository{}
				repo.On("IsConnected", mock.Anything, testID, applicationsTable).Return(false, nil)
				repo.On("UpdateStatus", mock.Anything, testID, applicationsTable).Return(testErr)
				return &repo
			},
			Request:          createRequestWithClaims(t, testID, consumer.Application),
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: "{\"errors\":[{\"message\":\"while updating status: test\"}]}\n",
			MockHandler:      testHandler(t, testID, consumer.Application),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepoFn()
			persist, transact := testCase.TxFn()
			update := statusupdate.New(transact, repo)

			// WHEN
			rr := httptest.NewRecorder()
			updateHandler := update.Handler()
			updateHandler(testCase.MockHandler).ServeHTTP(rr, testCase.Request)

			// THEN
			response := rr.Body.String()
			assert.Equal(t, testCase.ExpectedStatus, rr.Code)
			assert.Equal(t, testCase.ExpectedResponse, response)
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

func testHandler(t *testing.T, consumerID string, consumerType consumer.ConsumerType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cons, err := consumer.LoadFromContext(r.Context())
		require.NoError(t, err)

		require.Equal(t, consumerID, cons.ConsumerID)
		require.Equal(t, consumerType, cons.ConsumerType)

		_, err = w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}
