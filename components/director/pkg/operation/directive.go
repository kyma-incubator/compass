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

const ModeParam = "mode"

type directive struct {
	transact            persistence.Transactioner
	scheduler           Scheduler
	resourceFetcherFunc ResourceFetcherFunc
	tenantLoaderFunc    TenantLoaderFunc
}

// NewDirective creates a new handler struct responsible for the Async directive business logic
func NewDirective(transact persistence.Transactioner, scheduler Scheduler, resourceFetcherFunc ResourceFetcherFunc, tenantLoaderFunc TenantLoaderFunc) *directive {
	return &directive{
		transact:            transact,
		scheduler:           scheduler,
		resourceFetcherFunc: resourceFetcherFunc,
		tenantLoaderFunc:    tenantLoaderFunc,
	}
}

// HandleOperation enriches the request with an Operation information when the requesting mutation is annotated with the Async directive
func (d *directive) HandleOperation(ctx context.Context, _ interface{}, next gqlgen.Resolver, op graphql.OperationType, idField *string) (res interface{}, err error) {
	resCtx := gqlgen.GetResolverContext(ctx)
	var mode graphql.OperationMode
	if _, found := resCtx.Args[ModeParam]; !found {
		mode = graphql.OperationModeSync
	} else {
		modePointer, ok := resCtx.Args[ModeParam].(*graphql.OperationMode)
		if !ok {
			return nil, apperrors.NewInternalError(fmt.Sprintf("could not get %s parameter", ModeParam))
		}
		mode = *modePointer
	}

	ctx = SaveModeToContext(ctx, mode)

	tx, err := d.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to initialize database operation")
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := d.concurrencyCheck(ctx, op, resCtx, idField); err != nil {
		return nil, err
	}

	if mode == graphql.OperationModeSync {
		resp, err := next(ctx)
		if err != nil {
			return nil, err
		}

		err = tx.Commit()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
			return nil, apperrors.NewInternalError("Unable to finalize database operation")
		}

		return resp, nil
	}

	operation := &Operation{
		OperationType:     op,
		OperationCategory: resCtx.Field.Name,
		CorrelationID:     log.C(ctx).Data[log.FieldRequestID].(string),
	}
	ctx = SaveToContext(ctx, &[]*Operation{operation})

	resp, err := next(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while processing operation: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to process operation")
	}

	operationID, err := d.scheduler.Schedule(ctx, operation)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while scheduling operation: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to schedule operation")
	}

	operation.OperationID = operationID

	err = tx.Commit()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to finalize database operation")
	}

	return resp, nil
}

func (d *directive) concurrencyCheck(ctx context.Context, op graphql.OperationType, resCtx *gqlgen.ResolverContext, idField *string) error {
	if op == graphql.OperationTypeCreate {
		return nil
	}

	if idField == nil {
		return apperrors.NewInternalError("idField from context should not be empty")
	}

	resourceID, ok := resCtx.Args[*idField].(string)
	if !ok {
		return apperrors.NewInternalError(fmt.Sprintf("could not get idField: %q from request context", *idField))
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

	if app.DeletedAt.IsZero() && app.UpdatedAt.IsZero() && !app.Ready && (app.Error == nil || *app.Error == "") { // CREATING
		return apperrors.NewConcurrentOperationInProgressError("create operation is in progress")
	}
	if !app.DeletedAt.IsZero() && (app.Error == nil || *app.Error == "") { // DELETING
		return apperrors.NewConcurrentOperationInProgressError("delete operation is in progress")
	}
	// Note: This will be needed when there is async UPDATE supported
	// if app.DeletedAt.IsZero() && app.UpdatedAt.After(app.CreatedAt) && !app.Ready && *app.Error == "" { // UPDATING
	// 	return nil, apperrors.NewInvalidData	Error("another operation is in progress")
	// }

	return nil
}
