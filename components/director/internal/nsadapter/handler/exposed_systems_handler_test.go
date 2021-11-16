package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/handler"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/handler/automock"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/nsmodel"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	txautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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
	t.Run("failed to retrieve request body", func(t *testing.T) {

		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		endpoint := handler.NewHandler(nil, nil, nil)

		reader := Reader{}
		req := httptest.NewRequest(http.MethodPut, "/v1", &reader)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

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
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

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

	t.Run("failed due to missing report type", func(t *testing.T) {
		endpoint := handler.NewHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPut, "/v1", nil)

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "missing or invalid required report type query parameter",
			},
		})
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed due to unknown report type", func(t *testing.T) {
		endpoint := handler.NewHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPut, "/v1", nil)
		q := req.URL.Query()
		q.Add("reportType", "unknown")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "missing or invalid required report type query parameter",
			},
		})
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed while validating request body", func(t *testing.T) {
		bodyWithoutSubaccount := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"des\"\n        }\n      ]\n    }\n  ]\n}")

		endpoint := handler.NewHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPut, "/v1", bodyWithoutSubaccount)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

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
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(nil, errors.New("test"))
		defer transact.AssertExpectations(t)

		endpoint := handler.NewHandler(nil, nil, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

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
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		defer tx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil)
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return(nil, errors.New("test"))
		defer tntSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(nil, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

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

	t.Run("got error details when provided id is not a subaccount", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil)
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "customer"}}, nil)
		defer tntSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(nil, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Provided id is not subaccount",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("got error details when subaccount is not found", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil)
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{}, nil)
		defer tntSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(nil, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Subaccount not found",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	//Delta report tests
	t.Run("got error while upserting application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\",\n          \"systemNumber\": \"number\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		appInput := nsmodel.ToAppRegisterInput(nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}, "loc-id")

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, appInput).Return("", errors.New("error"))

		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("successfully upsert application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\",\n          \"systemNumber\": \"number\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		appInput := nsmodel.ToAppRegisterInput(nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}, "loc-id")

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, appInput).Return("success", nil)

		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to get application by subaccount, location ID and virtual host", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(nil, errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed to register application from template", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		nsmodel.Mappings = append(nsmodel.Mappings, systemfetcher.TemplateMapping{
			Name:        "",
			ID:          "ss",
			SourceKey:   []string{"description"},
			SourceValue: []string{"description"},
		})
		defer clearMappings()

		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "",
			},
			TemplateID: "ss",
		}
		appInput := nsmodel.ToAppRegisterInput(system, "loc-id")
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").
			Return(nil, nsmodel.NewSystemNotFoundError("fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, appInput, str.Ptr(system.TemplateID)).Return("", errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("successfully create application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		nsmodel.Mappings = append(nsmodel.Mappings, systemfetcher.TemplateMapping{
			Name:        "",
			ID:          "ss",
			SourceKey:   []string{"description"},
			SourceValue: []string{"description"},
		})
		defer clearMappings()

		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "",
			},
			TemplateID: "ss",
		}
		appInput := nsmodel.ToAppRegisterInput(system, "loc-id")
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").
			Return(nil, nsmodel.NewSystemNotFoundError("fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, appInput, str.Ptr(system.TemplateID)).Return("success", nil)
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to update application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		input := nsmodel.ToAppUpdateInput(nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		})
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed to set label applicationType", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}
		input := nsmodel.ToAppUpdateInput(system)
		label := &model.LabelInput{
			Key:        "applicationType",
			Value:      system.SystemType,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("failed to set label systemProtocol", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}
		input := nsmodel.ToAppUpdateInput(system)
		label := &model.LabelInput{
			Key:        "applicationType",
			Value:      system.SystemType,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		protocolLabel := &model.LabelInput{
			Key:        "systemProtocol",
			Value:      system.Protocol,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(errors.New("error")).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("successfully update system", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}
		input := nsmodel.ToAppUpdateInput(system)
		label := &model.LabelInput{
			Key:        "applicationType",
			Value:      system.SystemType,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		protocolLabel := &model.LabelInput{
			Key:        "systemProtocol",
			Value:      system.Protocol,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(nil).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to list by SCC", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Rollback").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("fail to mark system as unreachable", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: &model.Application{
				BaseEntity: &model.BaseEntity{ID: "id"},
			},
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("MarkAsUnreachable", mock.Anything, appWithLabel.App.ID).Return(errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		marshal, _ := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
						LocationId: "loc-id",
					},
				},
			},
		})
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(marshal))
	})

	t.Run("successfully mark system as unreachable", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: &model.Application{
				BaseEntity: &model.BaseEntity{ID: "id"},
			},
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("MarkAsUnreachable", mock.Anything, appWithLabel.App.ID).Return(nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("success when report type is delta and value is empty", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": []\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		defer transact.AssertExpectations(t)

		ids := make([]string, 0, 0)
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{}, nil)
		defer tntSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(nil, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "delta")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	//Full report tests
	t.Run("got error while upserting application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\",\n          \"systemNumber\": \"number\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		appInput := nsmodel.ToAppRegisterInput(nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}, "loc-id")

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, appInput).Return("", errors.New("error"))

		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully upsert application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\",\n          \"systemNumber\": \"number\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		appInput := nsmodel.ToAppRegisterInput(nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}, "loc-id")

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, appInput).Return("success", nil)

		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to get application by subaccount, location ID and virtual host", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(nil, errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to register application from template", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		nsmodel.Mappings = append(nsmodel.Mappings, systemfetcher.TemplateMapping{
			Name:        "",
			ID:          "ss",
			SourceKey:   []string{"description"},
			SourceValue: []string{"description"},
		})
		defer clearMappings()

		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "",
			},
			TemplateID: "ss",
		}
		appInput := nsmodel.ToAppRegisterInput(system, "loc-id")
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").
			Return(nil, nsmodel.NewSystemNotFoundError("fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, appInput, str.Ptr(system.TemplateID)).Return("", errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully create application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		nsmodel.Mappings = append(nsmodel.Mappings, systemfetcher.TemplateMapping{
			Name:        "",
			ID:          "ss",
			SourceKey:   []string{"description"},
			SourceValue: []string{"description"},
		})
		defer clearMappings()

		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "",
			},
			TemplateID: "ss",
		}
		appInput := nsmodel.ToAppRegisterInput(system, "loc-id")
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").
			Return(nil, nsmodel.NewSystemNotFoundError("fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, appInput, str.Ptr(system.TemplateID)).Return("success", nil)
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to update application", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		input := nsmodel.ToAppUpdateInput(nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		})
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to set label applicationType", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}
		input := nsmodel.ToAppUpdateInput(system)
		label := &model.LabelInput{
			Key:        "applicationType",
			Value:      system.SystemType,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to set label systemProtocol", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}
		input := nsmodel.ToAppUpdateInput(system)
		label := &model.LabelInput{
			Key:        "applicationType",
			Value:      system.SystemType,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		protocolLabel := &model.LabelInput{
			Key:        "systemProtocol",
			Value:      system.Protocol,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(errors.New("error")).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully update system", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"description\"\n        }\n      ]\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		application := model.Application{
			BaseEntity: &model.BaseEntity{
				ID: "id",
			},
		}
		system := nsmodel.System{
			SystemBase: nsmodel.SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "number",
			},
			TemplateID: "",
		}
		input := nsmodel.ToAppUpdateInput(system)
		label := &model.LabelInput{
			Key:        "applicationType",
			Value:      system.SystemType,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		protocolLabel := &model.LabelInput{
			Key:        "systemProtocol",
			Value:      system.Protocol,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}
		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: nil,
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSystem", mock.Anything, "fd4f2041-fa83-48e0-b292-ff515bb776f0", "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(nil).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to list by SCC", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Rollback").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return(nil, errors.New("error"))
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("fail to mark system as unreachable", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: &model.Application{
				BaseEntity: &model.BaseEntity{ID: "id"},
			},
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("MarkAsUnreachable", mock.Anything, appWithLabel.App.ID).Return(errors.New("error"))
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully mark system as unreachable", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		listSccsTx.Mock.On("Rollback").Return(nil) //used in listSCCs
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs

		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: &model.Application{
				BaseEntity: &model.BaseEntity{ID: "id"},
			},
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("MarkAsUnreachable", mock.Anything, appWithLabel.App.ID).Return(nil)
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return(nil, errors.New("error"))
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("success when there no unreachable SCCs", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Rollback").Return(nil)
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return(nil, errors.New("error"))
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return([]*model.SccMetadata{&model.SccMetadata{
			Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
			LocationId: "loc-id",
		}}, nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("success when there no unreachable SCCs", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)
		tx.Mock.On("Rollback").Return(nil)
		defer tx.AssertExpectations(t)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)
		defer listSccsTx.AssertExpectations(t)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()                                   // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once()                                   //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once()                           //used in listSCCs
		transact.Mock.On("Begin").Return(tx, nil).Once()                                   //used in listAppsBySCC
		transact.Mock.On("Begin").Return(tx, nil).Once()                                   //used in markAsUnreachable for unknown SCC
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Once() //used in markAsUnreachable for unknown SCC
		defer transact.AssertExpectations(t)

		ids := []string{"fd4f2041-fa83-48e0-b292-ff515bb776f0"}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: "fd4f2041-fa83-48e0-b292-ff515bb776f0", Type: "subaccount"}}, nil)
		defer tntSvc.AssertExpectations(t)

		labelFilters := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "loc-id")),
		}
		labelFilters2 := []*labelfilter.LabelFilter{
			labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"locationId\":%s}", "other-loc-id")),
		}
		appWithLabel := model.ApplicationWithLabel{
			App: &model.Application{
				BaseEntity: &model.BaseEntity{ID: "id"},
			},
			SccLabel: &model.Label{
				Value: "{\"LocationId\":\"loc-id\",\"Host\":\"127.0.0.1:8080\"}",
			},
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters).Return(nil, errors.New("error")).Once()
		appSvc.Mock.On("ListSCCs", mock.Anything, "scc").Return([]*model.SccMetadata{
			{
				Subaccount: "fd4f2041-fa83-48e0-b292-ff515bb776f0",
				LocationId: "loc-id",
			},
			{
				Subaccount: "marked-as-unreachable",
				LocationId: "other-loc-id",
			},
		}, nil)
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilters2).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil).Once()
		appSvc.Mock.On("MarkAsUnreachable", mock.Anything, appWithLabel.App.ID).Return(nil)
		defer appSvc.AssertExpectations(t)

		endpoint := handler.NewHandler(&appSvc, &tntSvc, &transact)

		req := httptest.NewRequest(http.MethodPut, "/v1", body)
		q := req.URL.Query()
		q.Add("reportType", "full")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
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

func clearMappings() {
	nsmodel.Mappings = nil
}
