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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)

type Delegator struct {
	operationManager  *process.ProvisionOperationManager
	avsConfig         Config
	clientHolder      *clientHolder
	operationsStorage storage.Operations
	configForModel    *configForModel
}

type configForModel struct {
	groupId  int64
	parentId int64
}

type avsNonSuccessResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func NewDelegator(avsConfig Config, operationsStorage storage.Operations) *Delegator {
	return &Delegator{
		operationManager:  process.NewProvisionOperationManager(operationsStorage),
		avsConfig:         avsConfig,
		clientHolder:      newClientHolder(avsConfig),
		operationsStorage: operationsStorage,
		configForModel: &configForModel{
			groupId:  avsConfig.GroupId,
			parentId: avsConfig.ParentId,
		},
	}
}

func (del *Delegator) CreateEvaluation(logger logrus.FieldLogger, operation internal.ProvisioningOperation, evalAssistant EvalAssistant, url string) (internal.ProvisioningOperation, time.Duration, error) {
	logger.Infof("starting the step avs internal id [%d] and avs external id [%d]", operation.Avs.AvsEvaluationInternalId, operation.Avs.AVSEvaluationExternalId)

	var updatedOperation internal.ProvisioningOperation
	d := 0 * time.Second

	if evalAssistant.IsAlreadyCreated(operation.Avs) {
		logger.Infof("step has already been finished previously")
		updatedOperation = operation
	} else {
		logger.Infof("making avs calls to create the Evaluation")
		evaluationObject, err := evalAssistant.CreateBasicEvaluationRequest(operation, del.configForModel, url)
		if err != nil {
			logger.Errorf("step failed with error %v", err)
			return operation, 5 * time.Second, nil
		}

		evalResp, err := del.postRequest(evaluationObject, logger)
		if err != nil {
			errMsg := fmt.Sprintf("post to avs failed with error %v", err)
			retryConfig := evalAssistant.provideRetryConfig()
			return del.operationManager.RetryOperation(operation, errMsg, retryConfig.retryInterval, retryConfig.maxTime, logger)
		}

		evalAssistant.SetEvalId(&operation.Avs, evalResp.Id)

		updatedOperation, d = del.operationManager.UpdateOperation(operation)
	}

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

func (del *Delegator) DeleteAvsEvaluation(deProvisioningOperation internal.DeprovisioningOperation, logger logrus.FieldLogger, assistant EvalAssistant) (internal.DeprovisioningOperation, error) {
	if assistant.IsAlreadyDeleted(deProvisioningOperation.Avs) {
		logger.Infof("Evaluations have been deleted previously")
		return deProvisioningOperation, nil
	}

	if err := del.tryDeleting(assistant, deProvisioningOperation.Avs, logger); err != nil {
		return deProvisioningOperation, err
	}

	assistant.markDeleted(&deProvisioningOperation.Avs)

	updatedDeProvisioningOp, err := del.operationsStorage.UpdateDeprovisioningOperation(deProvisioningOperation)
	if err != nil {
		return deProvisioningOperation, err
	}
	return *updatedDeProvisioningOp, nil
}

func (del *Delegator) tryDeleting(assistant EvalAssistant, lifecycleData internal.AvsLifecycleData, logger logrus.FieldLogger) error {
	evaluationId := assistant.GetEvaluationId(lifecycleData)
	err := del.removeReferenceFromParentEval(logger, evaluationId)
	if err != nil {
		logger.Errorf("error while deleting reference for evaluation %v", err)
		return err
	}

	err = del.deleteEvaluation(logger, evaluationId)
	if err != nil {
		logger.Errorf("error while deleting evaluation %v", err)
	}
	return err
}
func (del *Delegator) removeReferenceFromParentEval(logger logrus.FieldLogger, evaluationId int64) error {
	absoluteURL := fmt.Sprintf("%s/child/%d", appendId(del.avsConfig.ApiEndpoint, del.avsConfig.ParentId), evaluationId)
	return del.deleteRequest(absoluteURL, logger, inferIfReferenceIsGone)
}
func (del *Delegator) deleteEvaluation(logger logrus.FieldLogger, evaluationId int64) error {
	absoluteURL := appendId(del.avsConfig.ApiEndpoint, evaluationId)
	return del.deleteRequest(absoluteURL, logger, cannotInfer)

}

func (del *Delegator) deleteRequest(absoluteURL string, logger logrus.FieldLogger, inferDelete func(resp *http.Response, logger2 logrus.FieldLogger) bool) error {
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
		return errors.Wrap(err, "Error during DELETE call for url")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 404 {
		logger.Infof("Delete successful for url [%s]", absoluteURL)
		return nil
	} else if inferDelete(resp, logger) {
		return nil
	} else {
		msg := fmt.Sprintf("Got unexpected status [%d] while deleting for url [%s]", resp.StatusCode, absoluteURL)
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
func cannotInfer(resp *http.Response, logger logrus.FieldLogger) bool {
	return false
}

func inferIfReferenceIsGone(resp *http.Response, logger logrus.FieldLogger) bool {
	nonSuccessResp, err := deserializeNonSuccessAvsResponse(resp)
	if err != nil {
		return false
	}
	logger.Infof("Non Success avs response is %+v", nonSuccessResp)
	return strings.Contains(strings.ToLower(nonSuccessResp.Message), "does not contain subevaluation")
}

func deserializeNonSuccessAvsResponse(resp *http.Response) (*avsNonSuccessResp, error) {
	dec := json.NewDecoder(resp.Body)
	var responseObject avsNonSuccessResp
	err := dec.Decode(&responseObject)
	return &responseObject, err
}
