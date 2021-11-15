package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/handler"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/handler/automock"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	txautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Reader struct {
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return 0, errors.New("some error")
}

func TestHandler_ServeHTTP(t *testing.T) {

	// TODO: Add test cases.

	bodyWithoutSubaccount := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"des\"\n        }\n      ]\n    }\n  ]\n}")
	body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

	//func TestUpsertCustomer(t *testing.T) {
	t.Run("failed to retrieve request body", func(t *testing.T) {

		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		endpoint := handler.NewHandler(nil, nil, nil)

		reader := Reader{}
		req := httptest.NewRequest(http.MethodPut, "/v1", &reader)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "failed to retrieve request body",
			},
		})
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed to parse request body", func(t *testing.T) {
		endpoint := handler.NewHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPut, "/v1", nil)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "failed to parse request body",
			},
		})
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed while validating request body", func(t *testing.T) {
		endpoint := handler.NewHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPut, "/v1", bodyWithoutSubaccount)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "value: (subaccount: cannot be blank.).",
			},
		})
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed while opening transaction", func(t *testing.T) {
		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(nil, errors.New("test"))
		defer transact.AssertExpectations(t)

		endpoint := handler.NewHandler(nil, nil, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusInternalServerError,
				Message: "Update failed",
			},
		})
		Verify(t, resp, http.StatusInternalServerError, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed while listing tenants", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		defer tx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil)
		defer transact.AssertExpectations(t)


		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc:=automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return(nil,errors.New("test"))

		endpoint := handler.NewHandler(nil, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusInternalServerError,
				Message: "Update failed",
			},
		})
		Verify(t, resp, http.StatusInternalServerError, httputils.ContentTypeApplicationJSON, string(marshal))
	})
}

func Verify(t *testing.T, resp *http.Response,
	expectedStatusCode int, expectedContentType string, expectedBody string) {

	body, err := ioutil.ReadAll(resp.Body)
	respBody := strings.TrimSuffix(string(body), "\n")
	if nil != err {
		t.Fatalf("Failed to read the response body: %v", err)
	}

	if status := resp.StatusCode; status != expectedStatusCode {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expectedStatusCode)
	}

	if contentType := resp.Header.Get(httputils.HeaderContentType); contentType != expectedContentType {
		t.Errorf("the response contains unexpected content type: got %s want %s",
			contentType, expectedContentType)
	}

	if respBody != expectedBody {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'",
			respBody, expectedBody)
	}
}
