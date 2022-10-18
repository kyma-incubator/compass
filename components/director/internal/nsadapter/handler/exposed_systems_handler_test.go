package handler_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/handler/automock"
	txautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/handler"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/nsmodel"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
)

const (
	testSubaccount  = "fd4f2041-fa83-48e0-b292-ff515bb776f0"
	deltaReportType = "delta"
	fullReportType  = "full"
)

func TestHandler_ServeHTTP(t *testing.T) {
	appWithLabel := model.ApplicationWithLabel{
		App: &model.Application{
			BaseEntity: &model.BaseEntity{ID: "id"},
		},
		SccLabel: &model.Label{
			Value: map[string]interface{}{"LocationID": "loc-id", "Host": "127.0.0.1:8080"},
		},
	}

	appWithLabel2 := model.ApplicationWithLabel{
		App: &model.Application{
			BaseEntity: &model.BaseEntity{ID: "id"},
		},
		SccLabel: &model.Label{
			Value: map[string]interface{}{"LocationID": "loc-id-2", "Host": "127.0.0.1:8080"},
		},
	}

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
	label := &model.LabelInput{
		Key:        "systemType",
		Value:      system.SystemType,
		ObjectID:   application.ID,
		ObjectType: model.ApplicationLabelableObject,
	}

	labelFilter := labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"LocationID\":\"%s\", \"Subaccount\":\"%s\"}", "loc-id", testSubaccount))
	labelFilter2 := labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"LocationID\":\"%s\", \"Subaccount\":\"%s\"}", "loc-id-2", testSubaccount))

	protocolLabel := &model.LabelInput{
		Key:        "systemProtocol",
		Value:      system.Protocol,
		ObjectID:   application.ID,
		ObjectType: model.ApplicationLabelableObject,
	}

	body := "{" +
		"\"type\": \"notification-service\"," +
		"\"value\": [{" +
		"	\"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\"," +
		"	\"locationId\": \"loc-id\"," +
		"	\"exposedSystems\": [{" +
		"		\"protocol\": \"HTTP\"," +
		"		\"host\": \"127.0.0.1:8080\"," +
		"		\"type\": \"otherSAPsys\"," +
		"		\"status\": \"disabled\"," +
		"		\"description\": \"description\"" +
		"	}]\n    " +
		"}]}"
	bodyWithSystemNumber := "{" +
		"\"type\": \"notification-service\"," +
		"\"value\": [{" +
		"	\"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\"," +
		"   \"locationId\": \"loc-id\"," +
		"   \"exposedSystems\": [{" +
		"		\"protocol\": \"HTTP\"," +
		"       \"host\": \"127.0.0.1:8080\"," +
		"       \"type\": \"otherSAPsys\"," +
		"       \"status\": \"disabled\"," +
		"       \"description\": \"description\"," +
		"       \"systemNumber\": \"number\"" +
		"    }]" +
		"}]}"
	bodyWithoutExposedSystems := "{" +
		"\"type\": \"notification-service\"," +
		"\"value\": [{" +
		"	\"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\"," +
		"	\"locationId\": \"loc-id\"," +
		"	\"exposedSystems\": []" +
		"}]}"

	t.Run("failed to parse request body", func(t *testing.T) {
		endpoint := handler.NewHandler(nil, nil, nil, nil, nil)

		req := createReportSystemsRequest(nil, deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "failed to parse request body",
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed due to missing report type", func(t *testing.T) {
		endpoint := handler.NewHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPut, "/v1", nil)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "the query parameter 'reportType' is missing or invalid",
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed due to unknown report type", func(t *testing.T) {
		endpoint := handler.NewHandler(nil, nil, nil, nil, nil)

		req := createReportSystemsRequest(nil, "unknown")
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "the query parameter 'reportType' is missing or invalid",
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed while validating request body", func(t *testing.T) {
		bodyWithoutSubaccount := "{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": [\n        {\n          \"protocol\": \"HTTP\",\n          \"host\": \"127.0.0.1:8080\",\n          \"type\": \"otherSAPsys\",\n          \"status\": \"disabled\",\n          \"description\": \"des\"\n        }\n      ]\n    }\n  ]\n}"

		endpoint := handler.NewHandler(nil, nil, nil, nil, nil)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutSubaccount), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusBadRequest,
				Message: "value: (subaccount: cannot be blank.).",
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusBadRequest, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed while opening transaction", func(t *testing.T) {
		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(nil, errors.New("test"))
		defer transact.AssertExpectations(t)

		endpoint := handler.NewHandler(nil, nil, nil, nil, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusInternalServerError,
				Message: "Update failed",
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusInternalServerError, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed while listing tenants", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil)
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return(nil, errors.New("test"))
		defer mock.AssertExpectationsForObjects(t, tx, &transact, &tntSvc)

		endpoint := handler.NewHandler(nil, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.Error{
				Code:    http.StatusInternalServerError,
				Message: "Update failed",
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusInternalServerError, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("got error details when provided id is not a subaccount", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil)
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ExternalTenant: testSubaccount, Type: "customer"}}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, &transact, &tntSvc)

		endpoint := handler.NewHandler(nil, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Provided id is not subaccount",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("got error details when subaccount is not found", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil)
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true)

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, &transact, &tntSvc)

		endpoint := handler.NewHandler(nil, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Subaccount not found",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	//Delta report tests
	t.Run("got error while upserting application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, input).Return(errors.New("error"))

		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithSystemNumber), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("successfully upsert application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, input).Return(nil)

		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithSystemNumber), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to get application by subaccount, location ID and virtual host", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(nil, errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed to register application from template", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").
			Return(nil, errors.New("Object not found"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, input, str.Ptr("ss")).Return("", errors.New("error"))

		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("successfully create application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").
			Return(nil, errors.New("Object not found"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, input, str.Ptr("ss")).Return("success", nil)

		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to update application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

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

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed to set label systemType", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		input := nsmodel.ToAppUpdateInput(system)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()

		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("failed to set label systemProtocol", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		input := nsmodel.ToAppUpdateInput(system)
		protocolLabel := &model.LabelInput{
			Key:        "systemProtocol",
			Value:      system.Protocol,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(errors.New("error")).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("successfully update system", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		input := nsmodel.ToAppUpdateInput(system)
		protocolLabel := &model.LabelInput{
			Key:        "systemProtocol",
			Value:      system.Protocol,
			ObjectID:   application.ID,
			ObjectType: model.ApplicationLabelableObject,
		}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(nil).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to list by SCC", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Once()
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("fail to mark system as unreachable", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		unreachableInput := model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel.App.ID, unreachableInput).Return(errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		expectedBody, err := json.Marshal(httputil.ErrorResponse{
			Error: httputil.DetailedError{
				Code:    http.StatusOK,
				Message: "Update/create failed for some on-premise systems",
				Details: []httputil.Detail{
					{
						Message:    "Creation failed",
						Subaccount: testSubaccount,
						LocationID: "loc-id",
					},
				},
			},
		})
		require.NoError(t, err)
		Verify(t, resp, http.StatusOK, httputils.ContentTypeApplicationJSON, string(expectedBody))
	})

	t.Run("successfully mark system as unreachable", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		unreachableInput := model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel.App.ID, unreachableInput).Return(nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successful successfully mark system as unreachable with two sccs connected to one subaccount", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    },{\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id-2\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Times(3)
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount, testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		unreachableInput := model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter2).Return([]*model.ApplicationWithLabel{&appWithLabel2}, nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel.App.ID, unreachableInput).Return(nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel2.App.ID, unreachableInput).Return(nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(body, deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("success when report type is delta and value is empty", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": []\n}")

		endpoint := handler.NewHandler(nil, nil, nil, nil, nil)

		req := createReportSystemsRequest(body, deltaReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	//Full report tests
	t.Run("got error while upserting application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, input).Return(errors.New("error"))

		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithSystemNumber), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully upsert application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsert
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("Upsert", mock.Anything, input).Return(nil)

		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithSystemNumber), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to get application by subaccount, location ID and virtual host", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(nil, errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to register application from template", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").
			Return(nil, errors.New("Object not found"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, input, str.Ptr("ss")).Return("", errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully create application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		nsmodel.Mappings = append(nsmodel.Mappings, nsmodel.TemplateMapping{
			Name:        "",
			ID:          "ss",
			SourceKey:   []string{"description"},
			SourceValue: []string{"description"},
		})
		defer clearMappings()

		setMappings()
		defer clearMappings()

		appInputJSON := "app-input-json"
		applicationTemplate := &model.ApplicationTemplate{}

		appTemplateSvc := automock.ApplicationTemplateService{}
		appTemplateSvc.Mock.On("Get", mock.Anything, "ss").Return(applicationTemplate, nil)
		appTemplateSvc.Mock.On("PrepareApplicationCreateInputJSON", applicationTemplate, mock.Anything).Return(appInputJSON, nil)

		input := model.ApplicationRegisterInput{}
		appConverterSvc := automock.ApplicationConverter{}
		appConverterSvc.Mock.On("CreateInputJSONToModel", mock.Anything, appInputJSON).Return(input, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").
			Return(nil, errors.New("Object not found"))
		appSvc.Mock.On("CreateFromTemplate", mock.Anything, input, str.Ptr("ss")).Return("success", nil)
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appTemplateSvc, &appConverterSvc, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, &appConverterSvc, &appTemplateSvc, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to update application", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

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

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to set label systemType", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		input := nsmodel.ToAppUpdateInput(system)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(errors.New("error"))
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to set label systemProtocol", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		input := nsmodel.ToAppUpdateInput(system)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(errors.New("error")).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully update system", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for upsertSccSystems
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		input := nsmodel.ToAppUpdateInput(system)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("GetSccSystem", mock.Anything, testSubaccount, "loc-id", "127.0.0.1:8080").Return(&application, nil)
		appSvc.Mock.On("Update", mock.Anything, application.ID, input).Return(nil)
		appSvc.Mock.On("SetLabel", mock.Anything, label).Return(nil).Once()
		appSvc.Mock.On("SetLabel", mock.Anything, protocolLabel).Return(nil).Once()
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(body), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("failed to list by SCC", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once() // used for list tenants
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Once()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return(nil, errors.New("error"))
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("fail to mark system as unreachable", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		unreachableInput := model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel.App.ID, unreachableInput).Return(errors.New("error"))
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successfully mark system as unreachable", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Twice()
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		unreachableInput := model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel.App.ID, unreachableInput).Return(nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("successful successfully mark system as unreachable with two sccs connected to one subaccount", func(t *testing.T) {
		body := strings.NewReader("{\n  \"type\": \"notification-service\",\n  \"value\": [\n    {\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id\",\n      \"exposedSystems\": []\n    },{\n      \"subaccount\": \"fd4f2041-fa83-48e0-b292-ff515bb776f0\",\n      \"locationId\": \"loc-id-2\",\n      \"exposedSystems\": []\n    }\n  ]\n}")

		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list by scc
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for mark as unreachable
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Times(3)
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Times(3)

		ids := []string{testSubaccount, testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		unreachableInput := model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil)
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter2).Return([]*model.ApplicationWithLabel{&appWithLabel2}, nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel.App.ID, unreachableInput).Return(nil)
		appSvc.Mock.On("Update", mock.Anything, appWithLabel2.App.ID, unreachableInput).Return(nil)
		appSvc.Mock.On("ListSCCs", mock.Anything).Return(nil, errors.New("error"))
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(body, fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("success when there no unreachable SCCs", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()         // used for list tenants
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once() //used in listSCCs
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Once()
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Twice()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil)

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return(nil, errors.New("error"))
		appSvc.Mock.On("ListSCCs", mock.Anything).Return([]*model.SccMetadata{&model.SccMetadata{
			Subaccount: testSubaccount,
			LocationID: "loc-id",
		}}, nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})

	t.Run("success when there no unreachable SCCs", func(t *testing.T) {
		tx := &txautomock.PersistenceTx{}
		tx.Mock.On("Commit").Return(nil)

		listSccsTx := &txautomock.PersistenceTx{}
		listSccsTx.Mock.On("Commit").Return(nil)

		transact := txautomock.Transactioner{}
		transact.Mock.On("Begin").Return(tx, nil).Once()                                     //used for list tenants
		transact.Mock.On("Begin").Return(tx, nil).Once()                                     //used for getting template
		transact.Mock.On("Begin").Return(tx, nil).Once()                                     //list for mark unreachable
		transact.Mock.On("Begin").Return(listSccsTx, nil).Once()                             //used in listSCCs
		transact.Mock.On("Begin").Return(tx, nil).Once()                                     //used in listAppsBySCC
		transact.Mock.On("Begin").Return(tx, nil).Once()                                     //used in markAsUnreachable for unknown SCC
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, tx).Return(true).Times(5) //used in markAsUnreachable for unknown SCC
		transact.Mock.On("RollbackUnlessCommitted", mock.Anything, listSccsTx).Return(true).Once()

		ids := []string{testSubaccount}
		tntSvc := automock.TenantService{}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: testSubaccount, Type: "subaccount"}}, nil).Once()
		ids = []string{"marked-as-unreachable"}
		tntSvc.Mock.On("ListsByExternalIDs", mock.Anything, ids).Return([]*model.BusinessTenantMapping{{ID: "id", ExternalTenant: "marked-as-unreachable", Type: "subaccount"}}, nil).Once()

		unreachableInput := model.ApplicationUpdateInput{SystemStatus: str.Ptr("unreachable")}

		labelFilter2 := labelfilter.NewForKeyWithQuery("scc", fmt.Sprintf("{\"LocationID\":\"%s\", \"Subaccount\":\"%s\"}", "other-loc-id", "marked-as-unreachable"))

		appSvc := automock.ApplicationService{}
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter).Return(nil, errors.New("error")).Once()
		appSvc.Mock.On("ListSCCs", mock.Anything).Return([]*model.SccMetadata{
			{
				Subaccount: testSubaccount,
				LocationID: "loc-id",
			},
			{
				Subaccount: "marked-as-unreachable",
				LocationID: "other-loc-id",
			},
		}, nil)
		appSvc.Mock.On("ListBySCC", mock.Anything, labelFilter2).Return([]*model.ApplicationWithLabel{&appWithLabel}, nil).Once()
		appSvc.Mock.On("Update", mock.Anything, appWithLabel.App.ID, unreachableInput).Return(nil)
		defer mock.AssertExpectationsForObjects(t, tx, listSccsTx, &transact, &appSvc, &tntSvc)

		endpoint := handler.NewHandler(&appSvc, nil, nil, &tntSvc, &transact)

		req := createReportSystemsRequest(strings.NewReader(bodyWithoutExposedSystems), fullReportType)
		rec := httptest.NewRecorder()

		endpoint.ServeHTTP(rec, req)

		resp := rec.Result()
		Verify(t, resp, http.StatusNoContent, httputils.ContentTypeApplicationJSON, "{}")
	})
}

func Verify(t *testing.T, resp *http.Response, expectedStatusCode int, expectedContentType string, expectedBody string) {
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

func setMappings() {
	nsmodel.Mappings = append(nsmodel.Mappings, nsmodel.TemplateMapping{
		Name:        "",
		ID:          "ss",
		SourceKey:   []string{"type"},
		SourceValue: []string{"otherSAPsys"},
	})
}

func createReportSystemsRequest(body io.Reader, reportType string) *http.Request {
	req := httptest.NewRequest(http.MethodPut, "/v1", body)
	q := req.URL.Query()
	q.Add("reportType", reportType)
	req.URL.RawQuery = q.Encode()
	return req
}
