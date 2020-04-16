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

		handler := service.NewHandler(nil, nil, nil, logrus.New())
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

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(apperrors.WrongInput("test"))

		handler := service.NewHandler(nil, &mockValid, nil, logrus.New())
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusBadRequest)

		mockValid.AssertExpectations(t)
	})

	t.Run("Error when converting service details", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while converting service input: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValid, nil, logrus.New())
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValid.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Error when loading request context", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while requesting Request Context: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockContextProvider.On("ForRequest", mock.Anything).Return(service.RequestContext{AppID: "test"}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValid, &mockContextProvider, logrus.New())
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValid.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when creating service", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		expectedError := "while creating Service: test"

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("CreatePackage", mock.Anything, mock.Anything).Return("", testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValid, &mockContextProvider, logrus.New())
		handler.Create(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)

		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockValid.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)

		expectedError := ""

		req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("CreatePackage", mock.Anything, mock.Anything).Return("test", nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValid, &mockContextProvider, logrus.New())
		handler.Create(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusOK)

		mockValid.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
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
		mockConverter.On("GraphQLToServiceDetails", mock.Anything).Return(model.ServiceDetails{}, testErr)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New())
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
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
		mockConverter.On("GraphQLToServiceDetails", mock.Anything).Return(model.ServiceDetails{}, nil)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New())
		handler.Get(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusOK)
		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
		handler.List(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
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
		mockConverter.On("GraphQLToServiceDetails", mock.Anything).Return(model.ServiceDetails{}, testErr)

		handler := service.NewHandler(&mockConverter, nil, &mockContextProvider, logrus.New())
		handler.List(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)

		mockClient.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})
	t.Run("Success", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("ListPackages", mock.Anything, mock.Anything).Return([]*graphql.PackageExt{}, nil)

		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
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

		handler := service.NewHandler(nil, nil, nil, logrus.New())
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

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(apperrors.WrongInput("test"))

		handler := service.NewHandler(nil, &mockValid, nil, logrus.New())
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusBadRequest)
		mockValid.AssertExpectations(t)
	})

	t.Run("Error when converting service details", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while converting service input: test"

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValid, nil, logrus.New())
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValid.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
	})

	t.Run("Error when loading request context", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "while requesting Request Context: test"

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockContextProvider.On("ForRequest", mock.Anything).Return(service.RequestContext{AppID: "test"}, testErr)

		handler := service.NewHandler(&mockConverter, &mockValid, &mockContextProvider, logrus.New())
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValid.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
	})

	t.Run("Error when service not found", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		expectedError := "entity with ID  not found"

		req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, apperrors.NotFound("test"))
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValid, &mockContextProvider, logrus.New())
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusNotFound)
		mockValid.AssertExpectations(t)
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

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValid, &mockContextProvider, logrus.New())
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValid.AssertExpectations(t)
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

		mockValid := automock.Validator{}
		mockValid.On("Validate", mock.Anything).Return(nil)

		mockConverter := automock.Converter{}
		mockConverter.On("DetailsToGraphQLCreateInput", mock.Anything).Return(graphql.PackageCreateInput{}, nil)
		mockConverter.On("GraphQLCreateInputToUpdateInput", mock.Anything).Return(graphql.PackageUpdateInput{}, nil)

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("GetPackage", mock.Anything, mock.Anything).Return(graphql.PackageExt{}, nil)
		mockClient.On("UpdatePackage", mock.Anything, mock.Anything).Return(testErr)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(&mockConverter, &mockValid, &mockContextProvider, logrus.New())
		handler.Update(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockValid.AssertExpectations(t)
		mockConverter.AssertExpectations(t)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
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

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
		handler.Delete(w, req)

		resp := w.Result()
		assertErrorResponse(t, resp, expectedError, http.StatusInternalServerError)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(testServiceDetails)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodDelete, target, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockContextProvider := automock.RequestContextProvider{}
		mockClient := automock.DirectorClient{}
		mockClient.On("DeletePackage", mock.Anything).Return(nil)
		mockContextProvider.On("ForRequest", mock.Anything).
			Return(service.RequestContext{AppID: "test", DirectorClient: &mockClient}, nil)

		handler := service.NewHandler(nil, nil, &mockContextProvider, logrus.New())
		handler.Delete(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
		mockContextProvider.AssertExpectations(t)
		mockClient.AssertExpectations(t)
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
