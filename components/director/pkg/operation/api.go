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

	"github.com/gorilla/mux"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// ResourceIDParam missing godoc
const ResourceIDParam = "resource_id"

// ResourceTypeParam missing godoc
const ResourceTypeParam = "resource_type"

// ResourceFetcherFunc missing godoc
// ResourceFetcherFunc defines a function which fetches a particular resource by tenant and resource ID
type ResourceFetcherFunc func(ctx context.Context, tenantID, resourceID string) (model.Entity, error)

// TenantLoaderFunc missing godoc
// TenantLoaderFunc defines a function which fetches the tenant for a particular request
type TenantLoaderFunc func(ctx context.Context) (string, error)

type handler struct {
	transact            persistence.Transactioner
	resourceFetcherFunc ResourceFetcherFunc
	tenantLoaderFunc    TenantLoaderFunc
}

// NewHandler creates a new handler struct associated with the Operations API
func NewHandler(transact persistence.Transactioner, resourceFetcherFunc ResourceFetcherFunc, tenantLoaderFunc TenantLoaderFunc) *handler {
	return &handler{
		transact:            transact,
		resourceFetcherFunc: resourceFetcherFunc,
		tenantLoaderFunc:    tenantLoaderFunc,
	}
}

// ServeHTTP handles the Operations API requests
func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	tenantID, err := h.tenantLoaderFunc(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while retrieving tenant from context: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to determine tenant for request"), http.StatusInternalServerError)
		return
	}

	routeVariables := mux.Vars(request)
	resourceID := routeVariables[ResourceIDParam]
	resourceType := routeVariables[ResourceTypeParam]

	op := &Operation{
		ResourceID:   resourceID,
		ResourceType: resource.Type(resourceType),
	}

	log.C(ctx).Infof("Executing Operation API with resourceType: %s and resourceID: %s", op.ResourceType, op.ResourceID)

	if err := op.Validate(); err != nil {
		apperrors.WriteAppError(ctx, writer, apperrors.NewInvalidDataError("Unexpected resource type and/or GUID"), http.StatusBadRequest)
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

	res, err := h.resourceFetcherFunc(ctx, tenantID, op.ResourceID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while fetching resource from database: %s", err.Error())

		if apperrors.IsNotFoundError(err) {
			apperrors.WriteAppError(ctx, writer, apperrors.NewNotFoundErrorWithMessage(resource.Application, op.ResourceID,
				fmt.Sprintf("Operation for application with id %s not found", op.ResourceID)), http.StatusNotFound)
			return
		}

		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to execute database operation"), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
		apperrors.WriteAppError(ctx, writer, apperrors.NewInternalError("Unable to finalize database operation"), http.StatusInternalServerError)
		return
	}

	opResponse := buildLastOperation(res)

	err = json.NewEncoder(writer).Encode(opResponse)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding operation data: %v", err)
	}
}

func buildLastOperation(resource model.Entity) *OperationResponse {
	opResponse := &OperationResponse{
		Operation: &Operation{
			ResourceID:   resource.GetID(),
			ResourceType: resource.GetType(),
		},
		Error: resource.GetError(),
	}

	opResponse.initializeOperationType(resource)
	opResponse.initializeOperationStatus(resource)
	opResponse.initializeCreationTime(resource)

	return opResponse
}
