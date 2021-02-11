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
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/go-ozzo/ozzo-validation/is"

	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

type OperationRequest struct {
	OperationType graphql.OperationType `json:"operation_type,omitempty"`
	ResourceType  resource.Type         `json:"resource_type"`
	ResourceID    string                `json:"resource_id"`
	Error         string                `json:"error"`
}

// ResourceUpdaterFunc defines a function which updates a particular resource ready and error status
type ResourceUpdaterFunc func(ctx context.Context, id string, ready bool, errorMsg *string) error

type updateOperationHandler struct {
	transact             persistence.Transactioner
	resourceUpdaterFuncs map[resource.Type]ResourceUpdaterFunc
}

// NewUpdateOperationHandler creates a new handler struct to update resource by operation
func NewUpdateOperationHandler(transact persistence.Transactioner, resourceUpdaterFuncs map[resource.Type]ResourceUpdaterFunc) *updateOperationHandler {
	return &updateOperationHandler{
		transact:             transact,
		resourceUpdaterFuncs: resourceUpdaterFuncs,
	}
}

// ServeHTTP handles the Operations API requests
func (h *updateOperationHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if request.Method != http.MethodPost {
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	bytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to read request body"), http.StatusInternalServerError)
		return
	}
	defer func() {
		err := request.Body.Close()
		if err != nil {
			log.C(ctx).WithError(err).Error("Failed to close request body")
		}
	}()

	var operation OperationRequest
	if err := json.Unmarshal(bytes, &operation); err != nil {
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to decode body to JSON"), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateStruct(&operation,
		validation.Field(&operation.ResourceID, is.UUID),
		validation.Field(&operation.OperationType, validation.Required, validation.In(graphql.OperationTypeCreate, graphql.OperationTypeUpdate, graphql.OperationTypeDelete)),
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
	if err := resourceUpdaterFunc(ctx, operation.ResourceID, true, getError(operation.Error)); err != nil {
		if apperrors.IsNotFoundError(err) {
			apperrors.WriteAppError(ctx, writer, apperrors.NewNotFoundError(operation.ResourceType, operation.ResourceID), http.StatusNotFound)
			return
		}

		log.C(ctx).WithError(err).Errorf("While updating resource %s with id %s", operation.ResourceType, operation.ResourceID)
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to update resource %s with id %s", operation.ResourceType, operation.ResourceID), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to finalize database operation"), http.StatusInternalServerError)
		return
	}
}

func getError(errorMsg string) *string {
	if len(errorMsg) > 0 {
		return &errorMsg
	}

	return nil
}
