package avs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

type Delegator struct {
	operationManager *process.ProvisionOperationManager
	avsConfig        Config
	clientHolder     *clientHolder
}

func NewDelegator(avsConfig Config, operationsStorage storage.Operations) *Delegator {
	return &Delegator{
		operationManager: process.NewProvisionOperationManager(operationsStorage),
		avsConfig:        avsConfig,
		clientHolder:     newClientHolder(avsConfig),
	}
}

func (del *Delegator) DoRun(logger logrus.FieldLogger, operation internal.ProvisioningOperation, evalAssistant EvalAssistant, url string) (internal.ProvisioningOperation, time.Duration, error) {
	logger.Infof("starting the step avs internal id [%d] and avs external id [%d]", operation.AvsEvaluationInternalId, operation.AVSEvaluationExternalId)

	if evalAssistant.CheckIfAlreadyDone(operation) {
		logger.Infof("step has already been finished previously")
		return operation, 0, nil
	}

	evaluationObject, err := evalAssistant.CreateBasicEvaluationRequest(operation, url)
	if err != nil {
		logger.Errorf("step failed with error %v", err)
		return operation, 5 * time.Second, nil
	}

	evalResp, err := del.postRequest(evaluationObject, logger)
	if err != nil {
		logger.Errorf("post to avs failed with error %v", err)
		return operation, 30 * time.Second, nil
	}

	evalAssistant.SetEvalId(&operation, evalResp.Id)

	updatedOperation, d := del.operationManager.UpdateOperation(operation)

	evalAssistant.AppendOverrides(updatedOperation.InputCreator, updatedOperation.AvsEvaluationInternalId)

	return updatedOperation, d, nil
}

func (del *Delegator) postRequest(evaluationRequest *BasicEvaluationCreateRequest, logger logrus.FieldLogger) (*BasicEvaluationCreateResponse, error) {
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

	logger.Infof("Sending json body %s", string(objAsBytes))

	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Got unexpected status %d while creating internal evaluation", resp.StatusCode)
		logger.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	responseObject, err := deserialize(resp, err)
	if err != nil {
		return nil, err
	}

	return responseObject, nil
}

func deserialize(resp *http.Response, err error) (*BasicEvaluationCreateResponse, error) {
	dec := json.NewDecoder(resp.Body)
	var responseObject BasicEvaluationCreateResponse
	err = dec.Decode(&responseObject)
	return &responseObject, err
}
