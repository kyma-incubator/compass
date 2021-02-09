/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package operation_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/require"
)

const (
	tenantID   = "d6f1d2bb-62f5-4971-9efe-8af93d6528a7"
	resourceID = "46eb9542-8b18-4e4d-96d1-67f7e9675bb2"
)

func TestServeHTTP(t *testing.T) {

	t.Run("when tenant is missing it should return internal server error", func(t *testing.T) {
		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, context.Background())

		handler := operation.NewHandler(nil, nil, func(ctx context.Context) (string, error) {
			return "", mockedError()
		})
		handler.ServeHTTP(writer, req)

		require.Contains(t, writer.Body.String(), "Unable to determine tenant for request")
		require.Equal(t, http.StatusInternalServerError, writer.Code)
	})

	t.Run("when resourceID and resourceType are missing it should return bad request", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, ctx)

		handler := operation.NewHandler(nil, nil, loadTenantFunc)
		handler.ServeHTTP(writer, req)

		require.Contains(t, writer.Body.String(), "Unexpected resource type and/or GUID")
		require.Equal(t, http.StatusBadRequest, writer.Code)
	})

	t.Run("when resource ID is not GUID it should return bad request", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, ctx)

		queryValues := req.URL.Query()
		queryValues.Add(operation.ResourceIDParam, "123")

		req.URL.RawQuery = queryValues.Encode()

		handler := operation.NewHandler(nil, nil, loadTenantFunc)
		handler.ServeHTTP(writer, req)

		require.Contains(t, writer.Body.String(), "Unexpected resource type and/or GUID")
		require.Equal(t, http.StatusBadRequest, writer.Code)
	})

	t.Run("when resource type is not application it should return bad request", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, ctx)

		queryValues := req.URL.Query()
		queryValues.Add(operation.ResourceTypeParam, "runtime")

		req.URL.RawQuery = queryValues.Encode()

		handler := operation.NewHandler(nil, nil, loadTenantFunc)
		handler.ServeHTTP(writer, req)

		require.Contains(t, writer.Body.String(), "Unexpected resource type and/or GUID")
		require.Equal(t, http.StatusBadRequest, writer.Code)
	})

	t.Run("when transaction fails to begin it should return internal server error", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, ctx)

		queryValues := req.URL.Query()
		queryValues.Add(operation.ResourceIDParam, resourceID)
		queryValues.Add(operation.ResourceTypeParam, resource.Application.ToLower())

		req.URL.RawQuery = queryValues.Encode()

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewHandler(mockedTransactioner, nil, loadTenantFunc)
		handler.ServeHTTP(writer, req)

		require.Equal(t, http.StatusInternalServerError, writer.Code)
		require.Contains(t, writer.Body.String(), "Unable to establish connection with database")
	})

	t.Run("when resource fetcher func fails to fetch missing application resource it should return not found", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, ctx)

		queryValues := req.URL.Query()
		queryValues.Add(operation.ResourceIDParam, resourceID)
		queryValues.Add(operation.ResourceTypeParam, resource.Application.ToLower())

		req.URL.RawQuery = queryValues.Encode()

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewHandler(mockedTransactioner, func(_ context.Context, _, _ string) (*model.Application, error) {
			return nil, apperrors.NewNotFoundError(resource.Application, resourceID)
		}, loadTenantFunc)
		handler.ServeHTTP(writer, req)

		require.Equal(t, http.StatusNotFound, writer.Code)
		require.Contains(t, writer.Body.String(), "Object not found")
	})

	t.Run("when resource fetcher func fails to fetch application it should return internal server error", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, ctx)

		queryValues := req.URL.Query()
		queryValues.Add(operation.ResourceIDParam, resourceID)
		queryValues.Add(operation.ResourceTypeParam, resource.Application.ToLower())

		req.URL.RawQuery = queryValues.Encode()

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewHandler(mockedTransactioner, func(_ context.Context, _, _ string) (*model.Application, error) {
			return nil, mockedError()
		}, loadTenantFunc)
		handler.ServeHTTP(writer, req)

		require.Equal(t, http.StatusInternalServerError, writer.Code)
		require.Contains(t, writer.Body.String(), "Unable to execute database operation")
	})

	t.Run("when transaction fails to commit it should return internal server error ", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		writer := httptest.NewRecorder()
		req := fixEmptyRequest(t, ctx)

		queryValues := req.URL.Query()
		queryValues.Add(operation.ResourceIDParam, resourceID)
		queryValues.Add(operation.ResourceTypeParam, resource.Application.ToLower())

		req.URL.RawQuery = queryValues.Encode()

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewHandler(mockedTransactioner, func(_ context.Context, _, _ string) (*model.Application, error) {
			return nil, nil
		}, loadTenantFunc)
		handler.ServeHTTP(writer, req)

		require.Equal(t, http.StatusInternalServerError, writer.Code)
		require.Contains(t, writer.Body.String(), "Unable to finalize database operation")
	})

	t.Run("when application is successfully fetched it should return a respective operation", func(t *testing.T) {
		ctx := tenant.SaveToContext(context.Background(), tenantID, tenantID)

		req := fixEmptyRequest(t, ctx)

		queryValues := req.URL.Query()
		queryValues.Add(operation.ResourceIDParam, resourceID)
		queryValues.Add(operation.ResourceTypeParam, resource.Application.ToLower())

		req.URL.RawQuery = queryValues.Encode()

		mockedErr := mockedError().Error()
		now := time.Now()
		type testCase struct {
			Name             string
			Application      *model.Application
			ExpectedResponse operation.OperationResponse
		}

		cases := []testCase{
			{
				Name:        "Successful CREATE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: time.Time{}, DeletedAt: time.Time{}, Ready: true}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeCreate,
					},
					Status: operation.OperationStatusSucceeded,
				},
			},
			{
				Name:        "Successful UPDATE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: now.Add(1 * time.Minute), DeletedAt: time.Time{}, Ready: true}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeUpdate,
					},
					Status: operation.OperationStatusSucceeded,
				},
			},
			{
				Name:        "Successful DELETE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: now, DeletedAt: now.Add(1 * time.Minute), Ready: true}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeDelete,
					},
					Status: operation.OperationStatusSucceeded,
				},
			},
			{
				Name:        "In Progress CREATE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: time.Time{}, DeletedAt: time.Time{}, Ready: false}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeCreate,
					},
					Status: operation.OperationStatusInProgress,
				},
			},
			{
				Name:        "In Progress UPDATE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: now.Add(1 * time.Minute), DeletedAt: time.Time{}, Ready: false}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeUpdate,
					},
					Status: operation.OperationStatusInProgress,
				},
			},
			{
				Name:        "In Progress DELETE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: now, DeletedAt: now.Add(1 * time.Minute), Ready: false}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeDelete,
					},
					Status: operation.OperationStatusInProgress,
				},
			},
			{
				Name:        "Failed CREATE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: time.Time{}, DeletedAt: time.Time{}, Ready: false, Error: &mockedErr}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeCreate,
					},
					Status: operation.OperationStatusFailed,
					Error:  &mockedErr,
				},
			},
			{
				Name:        "Failed UPDATE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: now.Add(1 * time.Minute), DeletedAt: time.Time{}, Ready: false, Error: &mockedErr}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeUpdate,
					},
					Status: operation.OperationStatusFailed,
					Error:  &mockedErr,
				},
			},
			{
				Name:        "Failed DELETE Operation",
				Application: &model.Application{BaseEntity: &model.BaseEntity{CreatedAt: now, UpdatedAt: now, DeletedAt: now.Add(1 * time.Minute), Ready: false, Error: &mockedErr}},
				ExpectedResponse: operation.OperationResponse{
					Operation: &operation.Operation{
						ResourceID:    resourceID,
						ResourceType:  resource.Application.ToLower(),
						OperationType: graphql.OperationTypeDelete,
					},
					Status: operation.OperationStatusFailed,
					Error:  &mockedErr,
				},
			},
		}

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(len(cases))
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		for _, testCase := range cases {
			t.Run(testCase.Name, func(t *testing.T) {
				handler := operation.NewHandler(mockedTransactioner, func(_ context.Context, _, _ string) (*model.Application, error) {
					return testCase.Application, nil
				}, loadTenantFunc)

				writer := httptest.NewRecorder()
				handler.ServeHTTP(writer, req)

				expectedBody, err := json.Marshal(testCase.ExpectedResponse)
				require.NoError(t, err)

				require.Equal(t, http.StatusOK, writer.Code)
				require.Equal(t, string(expectedBody), strings.TrimSpace(writer.Body.String()))
			})
		}
	})

}

func fixEmptyRequest(t *testing.T, ctx context.Context) *http.Request {
	req, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	require.NoError(t, err)

	return req
}

func loadTenantFunc(_ context.Context) (string, error) {
	return tenantID, nil
}
