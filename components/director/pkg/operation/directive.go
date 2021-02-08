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

// Scheduler is responsible for scheduling any provided Operation entity for later processing
//go:generate mockery -name=Scheduler -output=automock -outpkg=automock -case=underscore
type Scheduler interface {
	Schedule(operation Operation) (string, error)
}

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
func (d *directive) HandleOperation(ctx context.Context, _ interface{}, next gqlgen.Resolver, op graphql.OperationType) (res interface{}, err error) {
	resCtx := gqlgen.GetResolverContext(ctx)
	mode, ok := resCtx.Args[ModeParam].(*graphql.OperationMode)
	if !ok {
		return nil, apperrors.NewInternalError(fmt.Sprintf("could not get %s parameter", ModeParam))
	}

	ctx = SaveModeToContext(ctx, *mode)

	tx, err := d.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to initialize database operation")
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if *mode == graphql.OperationModeSync {
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

	tenant, err := d.tenantLoaderFunc(ctx)
	if err != nil {
		return nil, apperrors.NewTenantRequiredError()
	}
	app, err := d.resourceFetcherFunc(ctx, tenant, operation.ResourceID)
	if err != nil {
		return nil, apperrors.NewInternalError("could not found resource with id %s", operation.ResourceID)
	}

	if app.DeletedAt.IsZero() && app.CreatedAt.Equal(app.UpdatedAt) && !app.Ready && *app.Error == "" { // CREATING
		return nil, apperrors.NewInvalidDataError("another operation is in progress")
	}
	if !app.DeletedAt.IsZero() && *app.Error == "" { // DELETING
		return nil, apperrors.NewInvalidDataError("another operation is in progress")
	}
	// Note: This will be needed when there is async UPDATE supported
	// if app.DeletedAt.IsZero() && app.UpdatedAt.After(app.CreatedAt) && !app.Ready && *app.Error == "" { // UPDATING
	// 	return nil, apperrors.NewInvalidDataError("another operation is in progress")
	// }

	operationID, err := d.scheduler.Schedule(*operation)
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
