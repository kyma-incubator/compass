package avs

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

const (
	DefinitionType   = "BASIC"
	interval         = 180
	timeout          = 30000
	contentCheck     = "SAP Kyma Runtime Monitoring"
	contentCheckType = "CONTAINS"
	threshold        = "30000"
	visibility       = "PUBLIC"
)

type BasicEvaluationCreateRequest struct {
	DefinitionType   string `json:"definition_type"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Service          string `json:"service"`
	URL              string `json:"url"`
	CheckType        string `json:"check_type"`
	Interval         int32  `json:"interval"`
	TesterAccessId   int64  `json:"tester_access_id"`
	Timeout          int    `json:"timeout"`
	ReadOnly         bool   `json:"read_only"`
	ContentCheck     string `json:"content_check"`
	ContentCheckType string `json:"content_check_type"`
	Threshold        string `json:"threshold"`
	GroupId          int64  `json:"group_id"`
	Visibility       string `json:"visibility"`
}

type BasicEvaluationCreateResponse struct {
	DefinitionType   string `json:"definition_type"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Service          string `json:"service"`
	URL              string `json:"url"`
	CheckType        string `json:"check_type"`
	Interval         int32  `json:"interval"`
	TesterAccessId   int64  `json:"tester_access_id"`
	Timeout          int    `json:"timeout"`
	ReadOnly         bool   `json:"read_only"`
	ContentCheck     string `json:"content_check"`
	ContentCheckType string `json:"content_check_type"`
	Threshold        int64  `json:"threshold"`
	GroupId          int64  `json:"group_id"`
	Visibility       string `json:"visibility"`

	DateCreated                int64    `json:"dateCreated"`
	DateChanged                int64    `json:"dateChanged"`
	Owner                      string   `json:"owner"`
	Status                     string   `json:"status"`
	Alerts                     []int    `json:"alerts"`
	Tags                       []string `json:"tags"`
	Id                         int64    `json:"id"`
	LegacyCheckId              int64    `json:"legacy_check_id"`
	InternalInterval           int64    `json:"internal_interval"`
	AuthType                   string   `json:"auth_type"`
	IndividualOutageEventsOnly bool     `json:"individual_outage_events_only"`
	IdOnTester                 string   `json:"id_on_tester"`
}

func newBasicEvaluationCreateRequest(operation internal.ProvisioningOperation, configurator ModelConfigurator, groupId int64, url string) (*BasicEvaluationCreateRequest, error) {
	provisionParams, err := operation.GetProvisioningParameters()
	if err != nil {
		return nil, err
	}

	beName, beDescription := generateNameAndDescription(provisionParams.ErsContext.GlobalAccountID,
		provisionParams.ErsContext.SubAccountID, provisionParams.Parameters.Name, DefinitionType, operation.InstanceID, configurator.ProvideSuffix())

	return &BasicEvaluationCreateRequest{
		DefinitionType:   DefinitionType,
		Name:             beName,
		Description:      beDescription,
		Service:          beName,
		URL:              url,
		CheckType:        configurator.ProvideCheckType(),
		Interval:         interval,
		TesterAccessId:   configurator.ProvideTesterAccessId(),
		Timeout:          timeout,
		ReadOnly:         false,
		ContentCheck:     contentCheck,
		ContentCheckType: contentCheckType,
		Threshold:        threshold,
		GroupId:          groupId,
		Visibility:       visibility,
	}, nil
}

func generateNameAndDescription(globalAccountId string, subAccountId string, name string, definitionType string, instanceId string, beType string) (string, string) {
	beName := fmt.Sprintf("K8S-Azure-Kyma-%s-%s-%s_%s", name, globalAccountId, subAccountId, beType)
	beDescription := fmt.Sprintf("%s %s evaluation for SAP Kyma Runtime for Global Account [%s], Subaccount [%s], instance name [%s] and instance id [%s]",
		definitionType, beType, globalAccountId, subAccountId, name, instanceId)
	return beName, beDescription
}
