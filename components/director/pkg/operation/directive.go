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

const modeParam = "mode"

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
	mode, ok := resCtx.Args[modeParam].(*graphql.OperationMode)
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not get %s parameter", modeParam))
	}

	ctx = SaveModeToContext(ctx, *mode)

	tx, err := d.transact.Begin()
	if err != nil {
		return nil, err
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
			return nil, err
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
		return nil, err
	}

	operationID, err := d.scheduler.Schedule(*operation)
	if err != nil {
		return nil, err
	}

	operation.OperationID = operationID

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
