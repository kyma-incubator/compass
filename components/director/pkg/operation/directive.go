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
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type directive struct {
	transact  persistence.Transactioner
	scheduler Scheduler
}

func NewDirective(transact persistence.Transactioner, scheduler Scheduler) *directive {
	return &directive{
		transact:  transact,
		scheduler: scheduler,
	}
}

func (d *directive) HandleOperation(ctx context.Context, _ interface{}, next gqlgen.Resolver, op graphql.OperationType) (res interface{}, err error) {
	resCtx := gqlgen.GetResolverContext(ctx)
	mode, ok := resCtx.Args["mode"].(graphql.OperationMode)
	if !ok {
		return nil, errors.New(fmt.Sprint("Could not get mode parameter"))
	}

	ctx = SaveModeToContext(ctx, mode)

	if mode == graphql.OperationModeSync {
		return next(ctx)
	}

	operation := Operation{
		OperationType:     op,
		OperationCategory: resCtx.Field.Name,
		CorrelationID:     log.C(ctx).Data[log.FieldRequestID].(string),
		RelatedResources:  make([]RelatedResource, 0),
	}
	ctx = SaveToContext(ctx, operation)

	tx, err := d.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	resp, err := next(ctx)
	if err != nil {
		return nil, err
	}

	operation, err = FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if len(operation.RelatedResources) == 0 {
		return nil, errors.New("related resources cannot be empty")
	}

	operation.ResourceID = operation.RelatedResources[0].ResourceID
	operation.ResourceType = operation.RelatedResources[0].ResourceType

	operationID, err := d.scheduler.Schedule(operation)
	if err != nil {
		return nil, err
	}

	operation.OperationID = operationID

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//log.C(ctx).Infof("Runtime with ID %s is in scenario with the owning application entity", runtimeID)
	return resp, nil
}
