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
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/go-ozzo/ozzo-validation/is"

	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
)

const ResourceIDParam = "resource_id"
const ResourceTypeParam = "resource_type"

type ResourceFetcherFunc func(ctx context.Context, tenant, id string) (*model.Application, error)

type handler struct {
	transact            persistence.Transactioner
	resourceFetcherFunc ResourceFetcherFunc
}

func NewHandler(transact persistence.Transactioner, resourceFetcherFunc ResourceFetcherFunc) handler {
	return handler{
		transact:            transact,
		resourceFetcherFunc: resourceFetcherFunc,
	}
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while retrieving tenant from context: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to determine tenant for request"), http.StatusInternalServerError)
		return
	}

	queryParams := request.URL.Query()

	inputParams := struct {
		ResourceID   string
		ResourceType string
	}{
		ResourceID:   queryParams.Get(ResourceIDParam),
		ResourceType: queryParams.Get(ResourceTypeParam),
	}

	log.C(ctx).Infof("Executing Operation API with resourceType: %s and resourceID: %s", inputParams.ResourceType, inputParams.ResourceID)

	if err := validation.ValidateStruct(&inputParams,
		validation.Field(&inputParams.ResourceID, is.UUID),
		validation.Field(&inputParams.ResourceType, validation.Required, validation.In(string(resource.Application)))); err != nil {
		http.Error(writer, fmt.Sprintf("Unexpected resource type and/or ID"), http.StatusBadRequest)
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

	app, err := h.resourceFetcherFunc(ctx, tenantID, inputParams.ResourceID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while fetching application from database: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to execute database operation"), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to finalize database operation"), http.StatusInternalServerError)
		return
	}

	opResponse := &OperationResponse{
		Operation: &Operation{
			ResourceID:   inputParams.ResourceID,
			ResourceType: inputParams.ResourceType,
		},
		Error: app.Error,
	}

	if !app.DeletedAt.IsZero() {
		opResponse.OperationType = graphql.OperationTypeDelete
	} else if app.CreatedAt != app.UpdatedAt {
		opResponse.OperationType = graphql.OperationTypeUpdate
	} else {
		opResponse.OperationType = graphql.OperationTypeCreate
	}

	if app.Ready {
		opResponse.Status = OperationStatusSucceeded
	} else if app.Error != nil {
		opResponse.Status = OperationStatusFailed
	} else {
		opResponse.Status = OperationStatusInProgress
	}

	err = json.NewEncoder(writer).Encode(opResponse)
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while encoding operation data")
	}
}
