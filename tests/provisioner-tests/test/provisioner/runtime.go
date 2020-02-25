package provisioner

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

//
//type TestRuntime struct {
//	log               *logrus.Entry
//	provisioningInput gqlschema.ProvisionRuntimeInput
//	runtimeID         string
//}
//
//func NewRuntime(provisioningInput gqlschema.ProvisionRuntimeInput) *TestRuntime {
//	testRuntime := TestRuntime{
//		provisioningInput: provisioningInput,
//	}
//	return &testRuntime
//}
//
//func (r *TestRuntime) Provision() (operationStatusID, runtimeID string, err error) {
//	r.log.Infof("Starting provisioning...")
//	currentOperationID, runtimeID, err := r.testSuite.ProvisionerClient.ProvisionRuntime(r.provisioningInput)
//	if err != nil {
//		r.log.Error(fmt.Sprintf("Error while provisioning Runtime: %s", err))
//		return "", "", errors.New(r.GetCurrentStatus())
//	}
//	r.log = r.log.WithField("RuntimeId", runtimeID)
//
//	r.LogStatus("Provisioning started.")
//	r.isRunning = true
//	return r.currentOperationID, r.runtimeID, nil
//}
//
//func (r *TestRuntime) GetOperationStatus(operationID string) (gqlschema.OperationStatus, error) {
//	r.LogStatus("Fetching Operation Status...")
//	operationStatus, err := r.testSuite.ProvisionerClient.RuntimeOperationStatus(operationID)
//	if err != nil {
//		r.LogStatus(fmt.Sprintf("Error while fetching Operation Status: %s", err))
//		return gqlschema.OperationStatus{}, errors.New(r.GetCurrentStatus())
//	}
//	r.LogStatus(fmt.Sprintf("%s: %s: %s", operationStatus.Operation, operationStatus.State, *operationStatus.Message))
//	return operationStatus, nil
//}
//
//func (r *TestRuntime) GetCurrentOperationStatus() (gqlschema.OperationStatus, error) {
//	return r.GetOperationStatus(r.currentOperationID)
//}
//
//func (r *TestRuntime) GetRuntimeStatus() (gqlschema.RuntimeStatus, error) {
//	r.LogStatus("Fetching Runtime Status...")
//	runtimeStatus, err := r.testSuite.ProvisionerClient.RuntimeStatus(r.runtimeID)
//	if err != nil {
//		r.LogStatus(fmt.Sprintf("Error while fetching Runtime Status: %s", err))
//		return gqlschema.RuntimeStatus{}, errors.New(r.GetCurrentStatus())
//	}
//	return runtimeStatus, nil
//}
//
//func (r *TestRuntime) Deprovision() (operationStatusID string, err error) {
//	r.LogStatus("Starting deprovisioning...")
//	r.currentOperationID, err = r.testSuite.ProvisionerClient.DeprovisionRuntime(r.runtimeID)
//	if err != nil {
//		r.LogStatus(fmt.Sprintf("Error while deprovisioning runtime: %s", err))
//		return "", errors.New(r.GetCurrentStatus())
//	}
//	r.LogStatus("Deprovisioning started.")
//	return r.currentOperationID, nil
//}
//
//func (r *TestRuntime) GetCurrentStatus() string {
//	return r.status[len(r.status)-1]
//}
//
//func (r *TestRuntime) StatusToString() string {
//	strStatus := ""
//	for _, state := range r.status {
//		strStatus += fmt.Sprintf("\t%s\n", state)
//	}
//	return strStatus
//}
