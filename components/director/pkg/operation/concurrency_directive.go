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
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type concurrencyDirective struct {
	transact            persistence.Transactioner
	tenantLoaderFunc    TenantLoaderFunc
	resourceFetcherFunc ResourceFetcherFunc
}

// NewConcurrencyDirective creates a new handler struct responsible for the Async directive business logic
func NewConcurrencyDirective(transact persistence.Transactioner, resourceFetcherFunc ResourceFetcherFunc, tenantLoaderFunc TenantLoaderFunc) *concurrencyDirective {
	return &concurrencyDirective{
		transact:            transact,
		tenantLoaderFunc:    tenantLoaderFunc,
		resourceFetcherFunc: resourceFetcherFunc,
	}
}

// ConcurrencyCheck enriches the request with an Operation information when the requesting mutation is annotated with the Async directive
func (d *concurrencyDirective) ConcurrencyCheck(ctx context.Context, _ interface{}, next gqlgen.Resolver, operationType graphql.OperationType, idField, parentIdField *string) (res interface{}, err error) {
	resCtx := gqlgen.GetResolverContext(ctx)

	tx, err := d.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to initialize database operation")
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := d.concurrencyCheck(ctx, operationType, resCtx, idField, parentIdField); err != nil {
		return nil, err
	}

	resp, err := next(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while processing operation: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to process operation")
	}

	err = tx.Commit()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to finalize database operation")
	}

	return resp, nil
}

func (d *concurrencyDirective) concurrencyCheck(ctx context.Context, op graphql.OperationType, resCtx *gqlgen.ResolverContext, idField, parentIdField *string) error {
	if op == graphql.OperationTypeCreate && parentIdField == nil {
		return nil
	}

	if idField == nil && parentIdField == nil {
		return apperrors.NewInternalError("idField or parentIdField from context should not be empty")
	}

	var resourceID string
	var ok bool
	if parentIdField != nil {
		resourceID, ok = resCtx.Args[*parentIdField].(string)
		if !ok {
			return apperrors.NewInternalError(fmt.Sprintf("could not get parentIdField: %q from request context", *parentIdField))
		}
	} else {
		resourceID, ok = resCtx.Args[*idField].(string)
		if !ok {
			return apperrors.NewInternalError(fmt.Sprintf("could not get idField: %q from request context", *idField))
		}
	}

	tenant, err := d.tenantLoaderFunc(ctx)
	if err != nil {
		return apperrors.NewTenantRequiredError()
	}

	app, err := d.resourceFetcherFunc(ctx, tenant, resourceID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return err
		}

		return apperrors.NewInternalError("failed to fetch resource with id %s", resourceID)
	}

	if app.GetDeletedAt().IsZero() && app.GetUpdatedAt().IsZero() && !app.GetReady() && (app.GetError() == nil || *app.GetError() == "") { // CREATING
		return apperrors.NewConcurrentOperationInProgressError("create operation is in progress")
	}
	if !app.GetDeletedAt().IsZero() && (app.GetError() == nil || *app.GetError() == "") { // DELETING
		return apperrors.NewConcurrentOperationInProgressError("delete operation is in progress")
	}
	// Note: This will be needed when there is async UPDATE supported
	// if app.DeletedAt.IsZero() && app.UpdatedAt.After(app.CreatedAt) && !app.Ready && *app.Error == "" { // UPDATING
	// 	return nil, apperrors.NewInvalidData	Error("another operation is in progress")
	// }

	return nil
}
