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

package operation

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/go-ozzo/ozzo-validation/v4/is"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// OperationRequest is the expected request body when updating certain operation status
type OperationRequest struct {
	OperationType     OperationType `json:"operation_type,omitempty"`
	ResourceType      resource.Type `json:"resource_type"`
	ResourceID        string        `json:"resource_id"`
	Error             string        `json:"error"`
	OperationCategory string        `json:"operation_category,omitempty"`
}

// ResourceUpdaterFunc defines a function which updates a particular resource ready and error status
type ResourceUpdaterFunc func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error

// ResourceDeleterFunc defines a function which deletes a particular resource by ID
type ResourceDeleterFunc func(ctx context.Context, id string) error

type updateOperationHandler struct {
	transact             persistence.Transactioner
	resourceUpdaterFuncs map[resource.Type]ResourceUpdaterFunc
	resourceDeleterFuncs map[resource.Type]ResourceDeleterFunc
}

type errResponse struct {
	err        error
	statusCode int
}

type operationError struct {
	Error string `json:"error"`
}

// OperationCategoryUnpairApplication Operation category for unpair application mutation. It's used determine the operation status
const OperationCategoryUnpairApplication = "unpairApplication"

// NewUpdateOperationHandler creates a new handler struct to update resource by operation
func NewUpdateOperationHandler(transact persistence.Transactioner, resourceUpdaterFuncs map[resource.Type]ResourceUpdaterFunc, resourceDeleterFuncs map[resource.Type]ResourceDeleterFunc) *updateOperationHandler {
	return &updateOperationHandler{
		transact:             transact,
		resourceUpdaterFuncs: resourceUpdaterFuncs,
		resourceDeleterFuncs: resourceDeleterFuncs,
	}
}

// ServeHTTP handles the Operations API requests
func (h *updateOperationHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if request.Method != http.MethodPut {
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	operation, errResp := operationRequestFromBody(ctx, request)
	if errResp != nil {
		apperrors.WriteAppError(ctx, writer, errResp.err, errResp.statusCode)
		return
	}

	if err := validation.ValidateStruct(operation,
		validation.Field(&operation.ResourceID, is.UUID),
		validation.Field(&operation.OperationType, validation.Required, validation.In(OperationTypeCreate, OperationTypeUpdate, OperationTypeDelete)),
		validation.Field(&operation.ResourceType, validation.Required, validation.In(resource.Application))); err != nil {
		apperrors.WriteAppError(ctx, writer, apperrors.NewInvalidDataError("Invalid operation properties: %s", err), http.StatusBadRequest)
		return
	}

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening db transaction: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to establish connection with database"), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	resourceUpdaterFunc := h.resourceUpdaterFuncs[operation.ResourceType]
	opError, err := stringifiedJSONError(operation.Error)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while marshalling operation error: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to marshal error"), http.StatusInternalServerError)
		return
	}

	appConditionStatus := determineApplicationFinalStatus(operation, opError)
	switch operation.OperationType {
	case OperationTypeCreate:
		fallthrough
	case OperationTypeUpdate:
		if err := resourceUpdaterFunc(ctx, operation.ResourceID, true, opError, appConditionStatus); err != nil {
			log.C(ctx).WithError(err).Errorf("While updating resource %s with id %s: %v", operation.ResourceType, operation.ResourceID, err)
			apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to update resource %s with id %s", operation.ResourceType, operation.ResourceID), http.StatusInternalServerError)
			return
		}
	case OperationTypeDelete:
		resourceDeleterFunc := h.resourceDeleterFuncs[operation.ResourceType]
		if operation.Error != "" {
			if err := resourceUpdaterFunc(ctx, operation.ResourceID, true, opError, appConditionStatus); err != nil {
				log.C(ctx).WithError(err).Errorf("While updating resource %s with id %s: %v", operation.ResourceType, operation.ResourceID, err)
				apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to update resource %s with id %s", operation.ResourceType, operation.ResourceID), http.StatusInternalServerError)
				return
			}
		} else {
			if err := resourceDeleterFunc(ctx, operation.ResourceID); err != nil {
				log.C(ctx).WithError(err).Errorf("While deleting resource %s with id %s: %v", operation.ResourceType, operation.ResourceID, err)
				apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to delete resource %s with id %s", operation.ResourceType, operation.ResourceID), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to finalize database operation"), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func operationRequestFromBody(ctx context.Context, request *http.Request) (*OperationRequest, *errResponse) {
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, &errResponse{apperrors.NewInternalError("Unable to read request body"), http.StatusInternalServerError}
	}

	defer func() {
		err := request.Body.Close()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Failed to close request body: %v", err)
		}
	}()

	var operation OperationRequest
	if err := json.Unmarshal(bytes, &operation); err != nil {
		return nil, &errResponse{apperrors.NewInternalError("Unable to decode body to JSON"), http.StatusBadRequest}
	}

	return &operation, nil
}

func stringifiedJSONError(errorMsg string) (*string, error) {
	if len(errorMsg) == 0 {
		return nil, nil
	}

	opErr := operationError{Error: errorMsg}
	bytesErr, err := json.Marshal(opErr)
	if err != nil {
		return nil, err
	}

	stringifiedErr := string(bytesErr)
	return &stringifiedErr, nil
}

func determineApplicationFinalStatus(op *OperationRequest, opError *string) model.ApplicationStatusCondition {
	appConditionStatus := model.ApplicationStatusConditionInitial
	switch op.OperationType {
	case OperationTypeCreate:
		appConditionStatus = model.ApplicationStatusConditionCreateSucceeded
		if opError != nil || str.PtrStrToStr(opError) != "" {
			appConditionStatus = model.ApplicationStatusConditionCreateFailed
		}
	case OperationTypeUpdate:
		if op.OperationCategory == OperationCategoryUnpairApplication {
			appConditionStatus = model.ApplicationStatusConditionInitial
			if opError != nil || str.PtrStrToStr(opError) != "" {
				appConditionStatus = model.ApplicationStatusConditionUnpairFailed
			}
		} else {
			appConditionStatus = model.ApplicationStatusConditionUpdateSucceeded
			if opError != nil || str.PtrStrToStr(opError) != "" {
				appConditionStatus = model.ApplicationStatusConditionUpdateFailed
			}
		}
	case OperationTypeDelete:
		appConditionStatus = model.ApplicationStatusConditionDeleteSucceeded
		if opError != nil || str.PtrStrToStr(opError) != "" {
			appConditionStatus = model.ApplicationStatusConditionDeleteFailed
		}
	}

	return appConditionStatus
}
