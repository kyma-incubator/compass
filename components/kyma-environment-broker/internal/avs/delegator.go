package avs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

const overrideKey = "avs_bridge.config.evaluations.cluster.id"

type delegator struct {
	operationManager *process.OperationManager
	avsConfig        Config
	clientHolder     *clientHolder
}

func newDelegator(avsConfig Config, operationsStorage storage.Operations) *delegator {
	return &delegator{
		operationManager: process.NewOperationManager(operationsStorage),
		avsConfig:        avsConfig,
		clientHolder:     newClientHolder(avsConfig),
	}
}

func (del *delegator) doRun(logger logrus.FieldLogger, stepName string, operation internal.ProvisioningOperation,
	modelSupplier func(provisioningOperation internal.ProvisioningOperation) (*basicEvaluationCreateRequest, error)) (internal.ProvisioningOperation, time.Duration, error) {
	logger.Infof("starting the step %s", stepName)

	if operation.AvsEvaluationInternalId != 0 {
		msg := fmt.Sprintf("%s has already been finished previously", stepName)
		return del.operationManager.OperationSucceeded(operation, msg)
	}

	evaluationObject, err := modelSupplier(operation)
	if err != nil {
		logger.Errorf("Step %s failed with error %v", stepName, err)
		return del.operationManager.OperationFailedWithError(operation, err)
	}

	evalResp, err := del.postRequest(evaluationObject, logger)
	if err != nil {
		return del.operationManager.OperationFailedWithError(operation, err)
	}

	operation.AvsEvaluationInternalId = evalResp.Id

	updatedOperation, d := del.operationManager.UpdateOperation(operation)

	updatedOperation.InputCreator.SetOverrides("avs-bridge", []*gqlschema.ConfigEntryInput{
		{
			Key:   overrideKey,
			Value: strconv.FormatInt(updatedOperation.AvsEvaluationInternalId, 10),
		},
	})

	return updatedOperation, d, nil
}

func (del *delegator) postRequest(evaluationRequest *basicEvaluationCreateRequest, logger logrus.FieldLogger) (*basicEvaluationCreateResponse, error) {
	objAsBytes, err := json.Marshal(evaluationRequest)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, del.avsConfig.ApiEndpoint, bytes.NewReader(objAsBytes))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	httpClient, err := del.clientHolder.getClient(logger)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Got unexpected status %del while creating internal evaluation", resp.StatusCode)
		logger.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	responseObject, err := deserialize(resp, err)
	if err != nil {
		return nil, err
	}

	return responseObject, nil
}

func deserialize(resp *http.Response, err error) (*basicEvaluationCreateResponse, error) {
	dec := json.NewDecoder(resp.Body)
	var responseObject basicEvaluationCreateResponse
	err = dec.Decode(&responseObject)
	return &responseObject, err
}
