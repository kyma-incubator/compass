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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

func TestUpdateOperationHandler(t *testing.T) {

	t.Run("when request method is not PUT it should return method not allowed", func(t *testing.T) {
		writer := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		require.NoError(t, err)

		handler := operation.NewUpdateOperationHandler(nil, nil, nil)
		handler.ServeHTTP(writer, req)

		require.Contains(t, writer.Body.String(), "Method not allowed")
		require.Equal(t, http.StatusMethodNotAllowed, writer.Code)
	})

	t.Run("when request body is not valid it should return bad request", func(t *testing.T) {
		writer := httptest.NewRecorder()
		reader := bytes.NewReader([]byte(`{"resource_id": 1}`))
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, "/", reader)
		require.NoError(t, err)

		handler := operation.NewUpdateOperationHandler(nil, nil, nil)
		handler.ServeHTTP(writer, req)

		require.Contains(t, writer.Body.String(), "Unable to decode body to JSON")
		require.Equal(t, http.StatusBadRequest, writer.Code)
	})

	t.Run("when required input body properties are missing it should return bad request", func(t *testing.T) {
		writer := httptest.NewRecorder()
		req := fixPostRequestWithBody(t, context.Background(), `{}`)

		handler := operation.NewUpdateOperationHandler(nil, nil, nil)
		handler.ServeHTTP(writer, req)

		require.Contains(t, writer.Body.String(), "Invalid operation properties")
		require.Equal(t, http.StatusBadRequest, writer.Code)
	})

	t.Run("when transaction fails to begin it should return internal server error", func(t *testing.T) {
		writer := httptest.NewRecorder()
		req := fixPostRequestWithBody(t, context.Background(), fmt.Sprintf(`{"resource_id": "%s", "resource_type": "%s", "operation_type": "%s"}`, resourceID, resource.Application, operation.OperationTypeCreate))

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewUpdateOperationHandler(mockedTransactioner, map[resource.Type]operation.ResourceUpdaterFunc{
			resource.Application: func(ctx context.Context, id string, ready bool, errorMsg *string, appConditionStatus model.ApplicationStatusCondition) error {
				return nil
			},
		}, nil)
		handler.ServeHTTP(writer, req)

		require.Equal(t, http.StatusInternalServerError, writer.Code)
		require.Contains(t, writer.Body.String(), "Unable to establish connection with database")
	})

	t.Run("when transaction fails to commit it should return internal server error", func(t *testing.T) {
		writer := httptest.NewRecorder()
		req := fixPostRequestWithBody(t, context.Background(), fmt.Sprintf(`{"resource_id": "%s", "resource_type": "%s", "operation_type": "%s"}`, resourceID, resource.Application, operation.OperationTypeCreate))

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(mockedError()).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewUpdateOperationHandler(mockedTransactioner, map[resource.Type]operation.ResourceUpdaterFunc{
			resource.Application: func(ctx context.Context, id string, ready bool, errorMsg *string, appConditionStatus model.ApplicationStatusCondition) error {
				return nil
			},
		}, nil)
		handler.ServeHTTP(writer, req)

		require.Equal(t, http.StatusInternalServerError, writer.Code)
		require.Contains(t, writer.Body.String(), "Unable to finalize database operation")
	})

	t.Run("when update handler fails on CREATE/UPDATE operation", func(t *testing.T) {
		writer := httptest.NewRecorder()
		req := fixPostRequestWithBody(t, context.Background(), fmt.Sprintf(`{"resource_id": "%s", "resource_type": "%s", "operation_type": "%s"}`, resourceID, resource.Application, operation.OperationTypeCreate))

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewUpdateOperationHandler(mockedTransactioner, map[resource.Type]operation.ResourceUpdaterFunc{
			resource.Application: func(ctx context.Context, id string, ready bool, errorMsg *string, appConditionStatus model.ApplicationStatusCondition) error {
				return errors.New("failed to update")
			},
		}, nil)
		handler.ServeHTTP(writer, req)

		mockedTx.AssertNotCalled(t, "Commit")
		require.Equal(t, http.StatusInternalServerError, writer.Code)
		require.Contains(t, writer.Body.String(), "Unable to update resource application with id")
	})

	t.Run("when delete handler fails on DELETE operation", func(t *testing.T) {
		writer := httptest.NewRecorder()
		req := fixPostRequestWithBody(t, context.Background(), fmt.Sprintf(`{"resource_id": "%s", "resource_type": "%s", "operation_type": "%s"}`, resourceID, resource.Application, operation.OperationTypeDelete))

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		handler := operation.NewUpdateOperationHandler(mockedTransactioner, nil, map[resource.Type]operation.ResourceDeleterFunc{
			resource.Application: func(ctx context.Context, id string) error {
				return errors.New("failed to delete")
			},
		})
		handler.ServeHTTP(writer, req)

		mockedTx.AssertNotCalled(t, "Commit")
		require.Equal(t, http.StatusInternalServerError, writer.Code)
		require.Contains(t, writer.Body.String(), "Unable to delete resource application with id")
	})

	t.Run("when operation has finished", func(t *testing.T) {
		type testCase struct {
			Name               string
			OperationType      operation.OperationType
			ExpectedError      string
			Ready              bool
			AppConditionStatus model.ApplicationStatusCondition
			UpdateCalled       int
			DeleteCalled       int
		}
		cases := []testCase{
			{
				Name:               "CREATE with error",
				OperationType:      operation.OperationTypeCreate,
				ExpectedError:      "operation failed",
				Ready:              true,
				AppConditionStatus: model.ApplicationStatusConditionCreateFailed,
				UpdateCalled:       1,
			},
			{
				Name:               "CREATE with NO error",
				OperationType:      operation.OperationTypeCreate,
				Ready:              true,
				AppConditionStatus: model.ApplicationStatusConditionCreateSucceeded,
				UpdateCalled:       1,
			},
			{
				Name:               "UPDATE with error",
				OperationType:      operation.OperationTypeUpdate,
				ExpectedError:      "operation UPDATE failed",
				Ready:              true,
				AppConditionStatus: model.ApplicationStatusConditionUpdateFailed,
				UpdateCalled:       1,
			},
			{
				Name:               "UPDATE with NO error",
				OperationType:      operation.OperationTypeUpdate,
				Ready:              true,
				AppConditionStatus: model.ApplicationStatusConditionUpdateSucceeded,
				UpdateCalled:       1,
			},
			{
				Name:               "DELETE with error",
				OperationType:      operation.OperationTypeDelete,
				ExpectedError:      "operation DELETE failed",
				Ready:              true,
				AppConditionStatus: model.ApplicationStatusConditionDeleteFailed,
				UpdateCalled:       1,
			},
			{
				Name:          "DELETE with NO error",
				OperationType: operation.OperationTypeDelete,
				DeleteCalled:  1,
			},
		}

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceedsMultipleTimes(len(cases))
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		for _, testCase := range cases {
			t.Run(testCase.Name, func(t *testing.T) {
				writer := httptest.NewRecorder()
				expectedErrorMsg := testCase.ExpectedError
				req := fixPostRequestWithBody(t, context.Background(), fmt.Sprintf(`{"resource_id": "%s", "resource_type": "%s", "operation_type": "%s", "error": "%s"}`, resourceID, resource.Application, testCase.OperationType, expectedErrorMsg))

				updateCalled := 0
				deleteCalled := 0
				handler := operation.NewUpdateOperationHandler(mockedTransactioner, map[resource.Type]operation.ResourceUpdaterFunc{
					resource.Application: func(ctx context.Context, id string, ready bool, errorMsg *string, appConditionStatus model.ApplicationStatusCondition) error {
						require.Equal(t, resourceID, id)
						require.Equal(t, testCase.Ready, ready)
						require.Equal(t, testCase.AppConditionStatus, appConditionStatus)
						if expectedErrorMsg == "" {
							require.Nil(t, errorMsg)
						} else {
							require.Equal(t, fmt.Sprintf("{\"error\":%q}", expectedErrorMsg), *errorMsg)
						}
						updateCalled++
						return nil
					},
				}, map[resource.Type]operation.ResourceDeleterFunc{
					resource.Application: func(ctx context.Context, id string) error {
						require.Equal(t, resourceID, id)
						deleteCalled++
						return nil
					},
				})

				handler.ServeHTTP(writer, req)
				require.Equal(t, testCase.UpdateCalled, updateCalled)
				require.Equal(t, testCase.DeleteCalled, deleteCalled)
				require.Equal(t, http.StatusOK, writer.Code)
			})
		}
	})
}

func fixPostRequestWithBody(t *testing.T, ctx context.Context, body string) *http.Request {
	reader := bytes.NewReader([]byte(body))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, "/", reader)
	require.NoError(t, err)

	return req
}
