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
	"math"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/header"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// ModeParam missing godoc
const ModeParam = "mode"

// WebhookFetcherFunc defines a function which fetches the webhooks for a specific resource ID
type WebhookFetcherFunc func(ctx context.Context, resourceID string) ([]*model.Webhook, error)

type directive struct {
	transact            persistence.Transactioner
	webhookFetcherFunc  WebhookFetcherFunc
	resourceFetcherFunc ResourceFetcherFunc
	resourceUpdaterFunc ResourceUpdaterFunc
	tenantLoaderFunc    TenantLoaderFunc
	scheduler           Scheduler
}

// NewDirective creates a new handler struct responsible for the Async directive business logic
func NewDirective(transact persistence.Transactioner, webhookFetcherFunc WebhookFetcherFunc, resourceFetcherFunc ResourceFetcherFunc, resourceUpdaterFunc ResourceUpdaterFunc, tenantLoaderFunc TenantLoaderFunc, scheduler Scheduler) *directive {
	return &directive{
		transact:            transact,
		webhookFetcherFunc:  webhookFetcherFunc,
		resourceFetcherFunc: resourceFetcherFunc,
		resourceUpdaterFunc: resourceUpdaterFunc,
		tenantLoaderFunc:    tenantLoaderFunc,
		scheduler:           scheduler,
	}
}

// HandleOperation enriches the request with an Operation information when the requesting mutation is annotated with the Async directive
func (d *directive) HandleOperation(ctx context.Context, _ interface{}, next gqlgen.Resolver, operationType graphql.OperationType, webhookType *graphql.WebhookType, idField *string) (res interface{}, err error) {
	resCtx := gqlgen.GetFieldContext(ctx)
	mode, err := getOperationMode(resCtx)
	if err != nil {
		return nil, err
	}

	ctx = SaveModeToContext(ctx, *mode)

	tx, err := d.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to initialize database operation")
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := d.concurrencyCheck(ctx, operationType, resCtx, idField); err != nil {
		return nil, err
	}

	if *mode == graphql.OperationModeSync {
		return executeSyncOperation(ctx, next, tx)
	}

	operation := &Operation{
		OperationType:     OperationType(str.Title(operationType.String())),
		OperationCategory: resCtx.Field.Name,
		CorrelationID:     log.C(ctx).Data[log.FieldRequestID].(string),
	}

	ctx = SaveToContext(ctx, &[]*Operation{operation})
	operationsArr, _ := FromCtx(ctx)

	committed := false
	defer func() {
		if !committed {
			lastIndex := int(math.Max(0, float64(len(*operationsArr)-1)))
			*operationsArr = (*operationsArr)[:lastIndex]
		}
	}()

	resp, err := next(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while processing operation: %s", err.Error())
		return nil, err
	}

	entity, ok := resp.(graphql.Entity)
	if !ok {
		log.C(ctx).WithError(err).Errorf("An error occurred while casting the response entity: %v", err)
		return nil, apperrors.NewInternalError("Failed to process operation")
	}

	appConditionStatus, err := determineApplicationInProgressStatus(operation)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("While determining the application status condition: %v", err)
		return nil, err
	}

	if err := d.resourceUpdaterFunc(ctx, entity.GetID(), false, nil, *appConditionStatus); err != nil {
		log.C(ctx).WithError(err).Errorf("While updating resource %s with id %s and status condition %v: %v", entity.GetType(), entity.GetID(), appConditionStatus, err)
		return nil, apperrors.NewInternalError("Unable to update resource %s with id %s", entity.GetType(), entity.GetID())
	}

	operation.ResourceID = entity.GetID()
	operation.ResourceType = entity.GetType()

	if webhookType != nil {
		webhookIDs, err := d.prepareWebhookIDs(ctx, err, operation, *webhookType)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while retrieving webhooks: %s", err.Error())
			return nil, apperrors.NewInternalError("Unable to retrieve webhooks")
		}

		operation.WebhookIDs = webhookIDs
	}

	requestObject, err := d.prepareRequestObject(ctx, err, resp)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while preparing request data: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to prepare webhook request data")
	}

	operation.RequestObject = requestObject

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
	committed = true

	return resp, nil
}

func (d *directive) concurrencyCheck(ctx context.Context, op graphql.OperationType, resCtx *gqlgen.FieldContext, idField *string) error {
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

	if app.GetDeletedAt().IsZero() && app.GetUpdatedAt().IsZero() && !app.GetReady() && hasNoErrors(app) { // CREATING
		return apperrors.NewConcurrentOperationInProgressError("create operation is in progress")
	}
	if !app.GetDeletedAt().IsZero() && hasNoErrors(app) { // DELETING
		return apperrors.NewConcurrentOperationInProgressError("delete operation is in progress")
	}

	if app.GetDeletedAt().IsZero() && app.GetUpdatedAt().After(app.GetCreatedAt()) && !app.GetReady() && hasNoErrors(app) { // UPDATING or UNPAIRING
		return apperrors.NewConcurrentOperationInProgressError("another operation is in progress")
	}

	return nil
}

func (d *directive) prepareRequestObject(ctx context.Context, err error, res interface{}) (string, error) {
	if err != nil {
		return "", err
	}

	tenantID, err := d.tenantLoaderFunc(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve tenant from request")
	}

	resource, ok := res.(webhook.Resource)
	if !ok {
		return "", errors.New("entity is not a webhook provider")
	}

	reqHeaders, ok := ctx.Value(header.ContextKey).(http.Header)
	if !ok {
		return "", errors.New("failed to retrieve request headers")
	}

	headers := make(map[string]string)
	for key, value := range reqHeaders {
		headers[key] = value[0]
	}

	requestObject := &webhook.ApplicationLifecycleWebhookRequestObject{
		Application: resource,
		TenantID:    tenantID,
		Headers:     headers,
	}

	data, err := json.Marshal(requestObject)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (d *directive) prepareWebhookIDs(ctx context.Context, err error, operation *Operation, webhookType graphql.WebhookType) ([]string, error) {
	if err != nil {
		return nil, err
	}

	webhooks, err := d.webhookFetcherFunc(ctx, operation.ResourceID)
	if err != nil {
		return nil, err
	}

	webhookIDs := make([]string, 0)
	for _, currWebhook := range webhooks {
		if graphql.WebhookType(currWebhook.Type) == webhookType {
			webhookIDs = append(webhookIDs, currWebhook.ID)
		}
	}

	if len(webhookIDs) > 1 {
		return nil, errors.New("multiple webhooks per operation are not supported")
	}

	return webhookIDs, nil
}

func getOperationMode(resCtx *gqlgen.FieldContext) (*graphql.OperationMode, error) {
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

	return &mode, nil
}

func executeSyncOperation(ctx context.Context, next gqlgen.Resolver, tx persistence.PersistenceTx) (interface{}, error) {
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

func hasNoErrors(app model.Entity) bool {
	return (app.GetError() == nil || *app.GetError() == "")
}

func determineApplicationInProgressStatus(op *Operation) (*model.ApplicationStatusCondition, error) {
	var appStatusCondition model.ApplicationStatusCondition
	switch op.OperationType {
	case OperationTypeCreate:
		appStatusCondition = model.ApplicationStatusConditionCreating
	case OperationTypeUpdate:
		if op.OperationCategory == OperationCategoryUnpairApplication {
			appStatusCondition = model.ApplicationStatusConditionUnpairing
		} else {
			appStatusCondition = model.ApplicationStatusConditionUpdating
		}
	case OperationTypeDelete:
		appStatusCondition = model.ApplicationStatusConditionDeleting
	default:
		return nil, apperrors.NewInvalidStatusCondition(resource.Application)
	}

	return &appStatusCondition, nil
}
