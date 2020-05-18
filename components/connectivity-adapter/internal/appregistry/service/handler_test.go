package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Create(t *testing.T) {

	target := "http://example.com/foo"
	testErr := errors.New("test")
	testServiceDetails := fixServiceDetails()

	t.Run("Error when unmarshalling input", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()
		expectedError := "while unmarshalling service: EOF"

		handler := service.NewHandler(nil, nil, nil, logrus.New(), nil)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusBadRequest)

	})

	t.Run("Error when validating input", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while validating input, test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(apperrors.WrongInput("test"))

		handler := service.NewHandler(nil, &mockValidator, nil, logrus.New(), nil)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusBadRequest)

		mockValidator.AssertExpectations(t)
	})

	t.Run("Error when converting service details", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while converting service input: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValidator, nil, logrus.New(), nil)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Error when loading request context", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while requesting Request Context: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockContextProvider.On("ForRequest", mock.Anything).Return(service.RequestContext{AppID: "test"}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), nil)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when creating service", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		expectedError := "while creating Service: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("CreatePackage", mock.Anything, mock.Anything).Return("", testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), nil)
		handler.Create(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)

		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error when listing legacy service references", func(t *testing.T) {
		body, err := json.Marshal(fixServiceDetailsWithIdentifier())
		require.NoError(t, err)
		labels := graphql.Labels{
			"foo": map[string]interface{}{
				"id":         "foo",
				"identifier": "foo",
			},
		}

		expectedError := "while listing legacy services for Application with ID 'test': test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", AppLabels: labels, DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ListServiceReferences", labels).Return(nil, testErr)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when legacy service identifier is not unique", func(t *testing.T) {
		body, err := json.Marshal(fixServiceDetailsWithIdentifier())
		require.NoError(t, err)
		labels := graphql.Labels{
			"foo": map[string]interface{}{
				"id":         "foo",
				"identifier": "foo",
			},
		}

		expectedError := "Service with Identifier Test already exists"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", AppLabels: labels, DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ListServiceReferences", labels).Return([]service.LegacyServiceReference{
			{
				ID:         "test",
				Identifier: "Test",
			},
		}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusConflict)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when writing legacy service reference", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		expectedError := "while setting Application label with legacy service metadata: while writing Application label: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("CreatePackage", mock.Anything, mock.Anything).Return("test", nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("WriteServiceReference", graphql.Labels(nil), service.LegacyServiceReference{
			ID:         "test",
			Identifier: "",
		}).Return(graphql.LabelInput{}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when setting Application label", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		expectedError := "while setting Application label with legacy service metadata: while setting Application label: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("CreatePackage", mock.Anything, mock.Anything).Return("test", nil)
		mockClient.On("SetApplicationLabel", "test", graphql.LabelInput{}).Return(testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("WriteServiceReference", graphql.Labels(nil), service.LegacyServiceReference{
			ID:         "test",
			Identifier: "",
		}).Return(graphql.LabelInput{}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		expectedError := ""

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("CreatePackage", mock.Anything, mock.Anything).Return("test", nil)
		mockClient.On("SetApplicationLabel", "test", graphql.LabelInput{}).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("WriteServiceReference", graphql.Labels(nil), service.LegacyServiceReference{
			ID:         "test",
			Identifier: "",
		}).Return(graphql.LabelInput{}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusOK)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Success when identifier specified", func(t *testing.T) {
		body, err := json.Marshal(fixServiceDetailsWithIdentifier())
		require.NoError(t, err)
		labels := graphql.Labels{
			"foo": map[string]interface{}{
				"id":         "foo",
				"identifier": "foo",
			},
		}

		expectedError := ""

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("CreatePackage", mock.Anything, mock.Anything).Return("test", nil)
		mockClient.On("SetApplicationLabel", "test", mock.Anything).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", AppLabels: labels, DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("WriteServiceReference", labels, service.LegacyServiceReference{
			ID:         "test",
			Identifier: "Test",
		}).Return(graphql.LabelInput{}, nil)
		mockLabeler.On("ListServiceReferences", labels).Return([]service.LegacyServiceReference{
			{
				ID:         "foo",
				Identifier: "foo",
			},
		}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusOK)

		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

}

func TestHandler_Get(t *testing.T) {

	target := "http://example.com/foo"
	testErr := errors.New("test")

	t.Run("Error when loading request context", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()
		expectedError := "while requesting Request Context: test"

		mockContextProvider := automock.RequestContextProvider{}
		mockContextProvider.On("ForRequest", mock.Anything).Return(service.RequestContext{AppID: "test"}, testErr)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when service not found", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		expectedError := "entity with ID  not found"

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, apperrors.NotFound("test"))

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusNotFound)

		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when fetching service", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		expectedError := "while fetching service: test"

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, testErr)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})
	t.Run("Error when converting service", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()
		expectedError := "while converting service: test"

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockConverter := automock.Converter{}
		mockConverter.On("GraphQLToServiceDetails", mock.Anything, mock.Anything).Return(model.ServiceDetails{}, testErr)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", graphql.Labels(nil), "").Return(service.LegacyServiceReference{}, nil)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when reading legacy service reference", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()
		expectedError := "while reading legacy service reference for Package with ID '': test"

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockConverter := automock.Converter{}

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", graphql.Labels(nil), "").Return(service.LegacyServiceReference{}, testErr)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()
		expectedError := ""

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockConverter := automock.Converter{}
		mockConverter.On("GraphQLToServiceDetails", mock.Anything, mock.Anything).Return(model.ServiceDetails{}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", graphql.Labels(nil), "").Return(service.LegacyServiceReference{}, nil)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusOK)
		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

}

func TestHandler_List(t *testing.T) {

	target := "http://example.com/foo"
	testErr := errors.New("test")

	t.Run("Error when loading request context", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()
		expectedError := "while requesting Request Context: test"

		mockContextProvider := automock.RequestContextProvider{}
		mockContextProvider.On("ForRequest", mock.Anything).Return(service.RequestContext{AppID: "test"}, testErr)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.List(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when fetching services", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		expectedError := "while fetching Services: test"

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("ListPackages", mock.Anything, mock.Anything).Return([]*graphql.PackageExt{}, testErr)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.List(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when reading legacy service references", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		expectedError := "while reading legacy service reference for Package with ID 'test': test"
		packages := []*graphql.PackageExt{{
			Package: graphql.Package{
				ID:   "test",
				Name: "test",
			},
		}}

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("ListPackages", mock.Anything, mock.Anything).Return(packages, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)
		mockConverter := automock.Converter{}

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", graphql.Labels(nil), "test").Return(service.LegacyServiceReference{}, testErr)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.List(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when converting service details to service", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		expectedError := "while converting detailed service to service: test"
		packages := []*graphql.PackageExt{{
			Package: graphql.Package{
				ID:   "test",
				Name: "test",
			},
		}}

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("ListPackages", mock.Anything, mock.Anything).Return(packages, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)
		mockConverter := automock.Converter{}
		mockConverter.On("GraphQLToServiceDetails", mock.Anything, mock.Anything).Return(model.ServiceDetails{}, nil)
		mockConverter.On("ServiceDetailsToService", mock.Anything, mock.Anything).Return(model.Service{}, testErr)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", graphql.Labels(nil), "test").Return(service.LegacyServiceReference{}, nil)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.List(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when converting services", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		expectedError := "while converting graphql to detailed service: test"
		packages := []*graphql.PackageExt{{
			Package: graphql.Package{
				ID:   "test",
				Name: "test",
			},
		}}

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("ListPackages", mock.Anything, mock.Anything).Return(packages, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)
		mockConverter := automock.Converter{}
		mockConverter.On("GraphQLToServiceDetails", mock.Anything, mock.Anything).Return(model.ServiceDetails{}, testErr)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", graphql.Labels(nil), "test").Return(service.LegacyServiceReference{}, nil)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.List(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("ListPackages", mock.Anything, mock.Anything).Return([]*graphql.PackageExt{}, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.List(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)
		assertContentTypeHeader(t, resp)
		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})
}

func TestHandler_Update(t *testing.T) {

	target := "http://example.com/foo"
	testErr := errors.New("test")
	testServiceDetails := fixServiceDetails()

	t.Run("Error when unmarshalling input", func(t *testing.T) {
		expectedError := "while unmarshalling service: EOF"

		req := httptest.NewRequest(http.MethodPut, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		handler := service.NewHandler(nil, nil, nil, logrus.New(), nil)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusBadRequest)
	})

	t.Run("Error when validating input", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while validating input, test"

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(apperrors.WrongInput("test"))

		handler := service.NewHandler(nil, &mockValidator, nil, logrus.New(), nil)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusBadRequest)
		mockValidator.AssertExpectations(t)
	})

	t.Run("Error when converting service details", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while converting service input: test"

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValidator, nil, logrus.New(), nil)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Error when loading request context", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while requesting Request Context: test"

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockContextProvider.On("ForRequest", mock.Anything).Return(service.RequestContext{AppID: "test"}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), nil)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when service not found", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "entity with ID  not found"

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, apperrors.NotFound("test"))
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), nil)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusNotFound)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error while fetching service", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := "while fetching service: test"

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), nil)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error when updating service", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := "while updating Service: test"

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)
		mockConverter.On("GraphQLCreateInputToUpdateInput", mock.Anything).Return(graphql.PackageUpdateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)
		mockClient.On("UpdatePackage", mock.Anything, mock.Anything).Return(testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), nil)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error when getting package", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := "while fetching service: test"

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)
		mockConverter.On("GraphQLCreateInputToUpdateInput", mock.Anything).Return(graphql.PackageUpdateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil).Once()
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, testErr).Once()
		mockClient.On("UpdatePackage", mock.Anything, mock.Anything).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when reading legacy service reference", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := "while reading legacy service reference for Package with ID '': test"

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)
		mockConverter.On("GraphQLCreateInputToUpdateInput", mock.Anything).Return(graphql.PackageUpdateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)
		mockClient.On("UpdatePackage", mock.Anything, mock.Anything).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", mock.Anything, mock.Anything).Return(service.LegacyServiceReference{}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when converting graphql to service details", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := "while converting service: test"

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)
		mockConverter.On("GraphQLCreateInputToUpdateInput", mock.Anything).Return(graphql.PackageUpdateInput{}, nil)
		mockConverter.On("GraphQLToServiceDetails", mock.Anything, mock.Anything).Return(model.ServiceDetails{}, testErr)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)
		mockClient.On("UpdatePackage", mock.Anything, mock.Anything).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", mock.Anything, mock.Anything).Return(service.LegacyServiceReference{}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := ""

		mockValidator := automock.Validator{}
		mockValidator.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)
		mockConverter.On("GraphQLCreateInputToUpdateInput", mock.Anything).Return(graphql.PackageUpdateInput{}, nil)
		mockConverter.On("GraphQLToServiceDetails", mock.Anything, mock.Anything).Return(model.ServiceDetails{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)
		mockClient.On("UpdatePackage", mock.Anything, mock.Anything).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("ReadServiceReference", mock.Anything, mock.Anything).Return(service.LegacyServiceReference{}, nil)

		handler := service.NewHandler(&mockConverter, &mockValidator, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusOK)
		mockValidator.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

}

func TestHandler_Delete(t *testing.T) {
	target := "http://example.com/foo"
	testErr := errors.New("test")
	testServiceDetails := fixServiceDetails()

	t.Run("Error when loading request context", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := "while requesting Request Context: test"

		mockContextProvider := automock.RequestContextProvider{}
		mockContextProvider.On("ForRequest", mock.Anything).Return(service.RequestContext{AppID: "test"}, testErr)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.Delete(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when service not found", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		expectedError := "entity with ID  not found"

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("DeletePackage", mock.Anything).Return(apperrors.NotFound("test"))
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.Delete(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusNotFound)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error when deleting service failed", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "test"

		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("DeletePackage", mock.Anything).Return(testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), nil)
		handler.Delete(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error when deleting legacy service reference", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while writing Application label: test"

		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("DeletePackage", mock.Anything).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("DeleteServiceReference", graphql.Labels(nil), "").Return(graphql.LabelInput{}, testErr)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Delete(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Error when setting Application label", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while setting Application label: test"

		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("DeletePackage", mock.Anything).Return(nil)
		mockClient.On("SetApplicationLabel", "test", graphql.LabelInput{}).Return(testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("DeleteServiceReference", graphql.Labels(nil), "").Return(graphql.LabelInput{}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Delete(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("DeletePackage", mock.Anything).Return(nil)
		mockClient.On("SetApplicationLabel", "test", graphql.LabelInput{}).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		mockLabeler := automock.AppLabeler{}
		mockLabeler.On("DeleteServiceReference", graphql.Labels(nil), "").Return(graphql.LabelInput{}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New(), &mockLabeler)
		handler.Delete(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLabeler.AssertExpectations(t)
	})

}

type response struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func getErrorMessage(t *testing.T, data []byte) string {
	var body response
	err := json.Unmarshal(data, &body)
	require.NoError(t, err)
	return body.Error
}

func assertErrorResponse(t *testing.T, resp *http.Response, expectedError string, expectedCode int) {
	require.NotNil(t, resp)
	assertContentTypeHeader(t, resp)

	respBody, err := ioutil.ReadAll(resp.Body)
	actualError := getErrorMessage(t, respBody)
	require.NoError(t, err)

	assert.Equal(t, expectedCode, resp.StatusCode)

	require.Equal(t, expectedError, actualError)

}

func assertContentTypeHeader(t *testing.T, resp *http.Response) {
	require.NotNil(t, resp)
	assert.Equal(t, res.HeaderContentTypeValue, resp.Header.Get(res.HeaderContentTypeKey))
}
