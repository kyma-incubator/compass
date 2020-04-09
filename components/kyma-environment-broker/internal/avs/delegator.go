package avs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

type Delegator struct {
	operationManager  *process.ProvisionOperationManager
	avsConfig         Config
	clientHolder      *clientHolder
	operationsStorage storage.Operations
}

func NewDelegator(avsConfig Config, operationsStorage storage.Operations) *Delegator {
	return &Delegator{
		operationManager:  process.NewProvisionOperationManager(operationsStorage),
		avsConfig:         avsConfig,
		clientHolder:      newClientHolder(avsConfig),
		operationsStorage: operationsStorage,
	}
}

func (del *Delegator) CreateEvaluation(logger logrus.FieldLogger, operation internal.ProvisioningOperation, evalAssistant EvalAssistant, url string) (internal.ProvisioningOperation, time.Duration, error) {
	logger.Infof("starting the step avs internal id [%d] and avs external id [%d]", operation.Avs.AvsEvaluationInternalId, operation.Avs.AVSEvaluationExternalId)

	if evalAssistant.IsAlreadyCreated(operation.Avs) {
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

	evalAssistant.SetEvalId(&operation.Avs, evalResp.Id)

	updatedOperation, d := del.operationManager.UpdateOperation(operation)

	evalAssistant.AppendOverrides(updatedOperation.InputCreator, updatedOperation.Avs.AvsEvaluationInternalId)

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Got unexpected status %d and response %s while creating evaluation", resp.StatusCode, responseBody(resp))
		logger.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	responseObject, err := deserializeCreateResponse(resp, err)
	if err != nil {
		return nil, err
	}

	return responseObject, nil
}

func deserializeCreateResponse(resp *http.Response, err error) (*BasicEvaluationCreateResponse, error) {
	dec := json.NewDecoder(resp.Body)
	var responseObject BasicEvaluationCreateResponse
	err = dec.Decode(&responseObject)
	return &responseObject, err
}

func responseBody(resp *http.Response) string {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

func (del *Delegator) DeleteAvsEvaluation(deProvisioningOperation internal.DeprovisioningOperation, logger logrus.FieldLogger, assistant EvalAssistant) (internal.DeprovisioningOperation, time.Duration, error) {
	if assistant.IsAlreadyDeleted(deProvisioningOperation.Avs) {
		logger.Infof("Evaluations have been deleted previously")
	}

	if err := del.tryDeleting(assistant, deProvisioningOperation.Avs, logger); err != nil {
		return deProvisioningOperation, time.Second * 10, nil
	}

	assistant.markDeleted(&deProvisioningOperation.Avs)

	updatedDeProvisioningOp, err := del.operationsStorage.UpdateDeprovisioningOperation(deProvisioningOperation)
	if err != nil {
		return deProvisioningOperation, time.Second * 10, nil
	}
	return *updatedDeProvisioningOp, 0, nil
}

func (del *Delegator) tryDeleting(assistant EvalAssistant, lifecycleData internal.AvsLifecycleData, logger logrus.FieldLogger) error {
	err := del.deleteRequest(logger, assistant.GetEvaluationId(lifecycleData))
	if err != nil {
		logger.Errorf("error while deleting evaluation %v", err)
	}
	return err
}

func (del *Delegator) deleteRequest(logger logrus.FieldLogger, evaluationId int64) error {
	absoluteURL := appendId(del.avsConfig.ApiEndpoint, evaluationId)

	req, err := http.NewRequest(http.MethodDelete, absoluteURL, nil)
	if err != nil {
		return err
	}

	httpClient, err := del.clientHolder.getClient(logger)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 404 {
		return nil
	} else {
		msg := fmt.Sprintf("Got unexpected status %d while deleting evaluation", resp.StatusCode)
		logger.Error(msg)
		return fmt.Errorf(msg)
	}

}

func appendId(baseUrl string, id int64) string {
	if strings.HasSuffix(baseUrl, "/") {
		return baseUrl + strconv.FormatInt(id, 10)
	} else {
		return baseUrl + "/" + strconv.FormatInt(id, 10)
	}
}
