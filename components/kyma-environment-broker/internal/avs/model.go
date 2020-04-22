package avs

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
)

const (
	DefinitionType   = "BASIC"
	interval         = 180
	timeout          = 30000
	contentCheck     = "error"
	contentCheckType = "NOT_CONTAINS"
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
	ParentId         int64  `json:"parent_id"`
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

func newBasicEvaluationCreateRequest(operation internal.ProvisioningOperation, evalTypeSpecificConfig ModelConfigurator,
	configForModel *configForModel, url string) (*BasicEvaluationCreateRequest, error) {
	provisionParams, err := operation.GetProvisioningParameters()
	if err != nil {
		return nil, err
	}

	beName, beDescription := generateNameAndDescription(provisionParams.ErsContext.GlobalAccountID,
		provisionParams.ErsContext.SubAccountID, provisionParams.Parameters.Name, evalTypeSpecificConfig.ProvideSuffix())

	return &BasicEvaluationCreateRequest{
		DefinitionType:   DefinitionType,
		Name:             beName,
		Description:      beDescription,
		Service:          beName,
		URL:              url,
		CheckType:        evalTypeSpecificConfig.ProvideCheckType(),
		Interval:         interval,
		TesterAccessId:   evalTypeSpecificConfig.ProvideTesterAccessId(),
		Timeout:          timeout,
		ReadOnly:         false,
		ContentCheck:     contentCheck,
		ContentCheckType: contentCheckType,
		Threshold:        threshold,
		GroupId:          configForModel.groupId,
		Visibility:       visibility,
		ParentId:         configForModel.parentId,
	}, nil
}

func generateNameAndDescription(globalAccountId string, subAccountId string, name string, beType string) (string, string) {
	beName := fmt.Sprintf("K8S-AZR-Kyma-%s-%s-%s", beType, subAccountId, name)
	beDescription := fmt.Sprintf("SKR instance Name: %s, Global Account: %s, Subaccount: %s",
		name, globalAccountId, subAccountId)

	return truncateString(beName, 80), truncateString(beDescription, 255)
}

func truncateString(input string, num int) string {
	output := input
	if len(input) > num {
		output = input[0:num]
	}
	return output
}
