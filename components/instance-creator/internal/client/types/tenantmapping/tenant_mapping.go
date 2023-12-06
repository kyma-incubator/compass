package tenantmapping

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"strings"
)

// Contract
// {
//  "state": "CONFIG_PENDING",
//  "configuration": {
//    "credentials": {
//      "inboundCommunication": {
//        "oauth2mtls": {
//          "tokenServiceUrl": "$.serviceInstances[1].serviceKey.url",
//          "clientId": "$.serviceInstances[1].serviceKey.clientid",
//          "certificate": "-----BEGIN CERTIFICATE----- certFromS4 -----END CERTIFICATE-----",
//          "correlationIds": ["SAP_COM_0545"],
//          "serviceInstances": [
//            {
//              "service": "procurement-service-test",
//              "plan": "apiaccess",
//              "configuration": {},
//              "serviceKey": {
//                "configuration": {}
//              }
//            },
//            {
//              "service": "identity",
//              "plan": "application",
//              "configuration": {
//                "consumed-services": [
//                  {
//                    "service-instance-name": "$.serviceInstances[0].name"
//                  }
//                ],
//                "xsuaa-cross-consumption": true
//              },
//              "serviceKey": {
//                "configuration": {
//                  "credential-type": "X509_PROVIDED",
//                  "certificate": "-----BEGIN CERTIFICATE----- certFromS4 -----END CERTIFICATE-----"
//                }
//              }
//            }
//          ]
//        }
//      }
//    }
//  }
//}

const (
	assignOperation   = "assign"
	unassignOperation = "unassign"

	inboundCommunicationKey  = "inboundCommunication"
	outboundCommunicationKey = "outboundCommunication"
)

// Context is a structure used to JSON decode the context in the Body
type Context struct {
	FormationID string `json:"uclFormationId"`
	Operation   string `json:"operation"`
}

// ReceiverTenant is a structure used to JSON decode the receiverTenant in the Body
type ReceiverTenant struct {
	Region        string          `json:"deploymentRegion"`
	SubaccountID  string          `json:"subaccountId"`
	Configuration json.RawMessage `json:"configuration"`
}

// AssignedTenant is a structure used to JSON decode the assignedTenant in the Body
type AssignedTenant struct {
	AssignmentID  string          `json:"uclAssignmentId"`
	Configuration json.RawMessage `json:"configuration"`
}

// Body is a structure used to JSON decode the request body sent to the adapter handler
type Body struct {
	Context        Context        `json:"context"`
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	AssignedTenant AssignedTenant `json:"assignedTenant"`
}

// GetAssignedTenantInboundCommunication returns the Body assigned tenant inbound communication
func (b Body) GetAssignedTenantInboundCommunication() gjson.Result {
	assignedTenantConfiguration := gjson.ParseBytes(b.AssignedTenant.Configuration)

	assignedTenantInboundCommunicationPath := FindKeyPath(assignedTenantConfiguration.Value(), inboundCommunicationKey)
	if assignedTenantInboundCommunicationPath == "" {
		return gjson.Result{}
	}

	assignedTenantInboundCommunication := gjson.GetBytes(b.AssignedTenant.Configuration, assignedTenantInboundCommunicationPath)
	if !assignedTenantInboundCommunication.Exists() {
		return gjson.Result{}
	}

	return assignedTenantInboundCommunication
}

// GetReceiverTenantOutboundCommunication returns the Body receiver tenant outbound communication
func (b Body) GetReceiverTenantOutboundCommunication() gjson.Result {
	receiverTenantConfiguration := gjson.ParseBytes(b.ReceiverTenant.Configuration)

	receiverTenantOutboundCommunicationPath := FindKeyPath(receiverTenantConfiguration.Value(), outboundCommunicationKey)
	if receiverTenantOutboundCommunicationPath == "" {
		return gjson.Result{}
	}

	receiverTenantOutboundCommunication := gjson.GetBytes(b.ReceiverTenant.Configuration, receiverTenantOutboundCommunicationPath)
	if !receiverTenantOutboundCommunication.Exists() {
		return gjson.Result{}
	}

	return receiverTenantOutboundCommunication
}

// GetReceiverTenantInboundCommunication returns the Body receiver tenant inbound communication
func (b Body) GetReceiverTenantInboundCommunication() gjson.Result {
	receiverTenantConfiguration := gjson.ParseBytes(b.ReceiverTenant.Configuration)

	receiverTenantInboundCommunicationPath := FindKeyPath(receiverTenantConfiguration.Value(), inboundCommunicationKey)
	if receiverTenantInboundCommunicationPath == "" {
		return gjson.Result{}
	}

	receiverTenantInboundCommunication := gjson.GetBytes(b.ReceiverTenant.Configuration, receiverTenantInboundCommunicationPath)
	if !receiverTenantInboundCommunication.Exists() {
		return gjson.Result{}
	}

	return receiverTenantInboundCommunication
}

func (b *Body) SetReceiverTenantAuth(authKey string, authValue map[string]interface{}) error {
	receiverTenantConfiguration := gjson.ParseBytes(b.ReceiverTenant.Configuration)

	receiverTenantOutboundCommunicationPath := FindKeyPath(receiverTenantConfiguration.Value(), outboundCommunicationKey)
	if receiverTenantOutboundCommunicationPath == "" {
		return errors.New("Receiver tenant inbound communication is missing in the configuration")
	}

	newReceiverTenantConfiguration, err := sjson.SetBytes(b.ReceiverTenant.Configuration, fmt.Sprintf("%s.%s", receiverTenantOutboundCommunicationPath, authKey), authValue)
	if err != nil {
		return errors.Wrapf(err, "while setting receiver tenant %q auth key in outbound communication", authKey)
	}
	b.ReceiverTenant.Configuration = newReceiverTenantConfiguration
	return nil
}

// Validate validates the request Body
func (b Body) Validate() error {
	if b.Context.FormationID == "" {
		return apperrors.NewInvalidDataError("Context's Formation ID should be provided")
	}

	if b.Context.Operation == "" || (b.Context.Operation != assignOperation && b.Context.Operation != unassignOperation) {
		return apperrors.NewInvalidDataError("Context's Operation is invalid, expected %q or %q, got: %q", assignOperation, unassignOperation, b.Context.Operation)
	}

	if b.AssignedTenant.AssignmentID == "" {
		return apperrors.NewInvalidDataError("Assigned Tenant Assignment ID should be provided")
	}

	if b.ReceiverTenant.Region == "" {
		return apperrors.NewInvalidDataError("Receiver Tenant Region should be provided")
	}

	if b.ReceiverTenant.SubaccountID == "" {
		return apperrors.NewInvalidDataError("Receiver Tenant Subaccount ID should be provided")
	}

	if b.Context.Operation == assignOperation {
		assignedTenantConfiguration := gjson.ParseBytes(b.AssignedTenant.Configuration)
		assignedTenantInboundCommunicationPath := FindKeyPath(assignedTenantConfiguration.Value(), inboundCommunicationKey)
		if assignedTenantInboundCommunicationPath == "" {
			return apperrors.NewInvalidDataError("Assigned tenant inbound communication is missing in the configuration")
		}

		receiverTenantConfiguration := gjson.ParseBytes(b.ReceiverTenant.Configuration)
		receiverTenantOutboundCommunicationPath := FindKeyPath(receiverTenantConfiguration.Value(), outboundCommunicationKey)
		if receiverTenantOutboundCommunicationPath != "" && strings.TrimSuffix(receiverTenantOutboundCommunicationPath, outboundCommunicationKey) != strings.TrimSuffix(assignedTenantInboundCommunicationPath, inboundCommunicationKey) {
			return errors.New("Receiver tenant outbound communication should be in the same place as the assigned tenant inbound communication")
		}
	}

	return nil
}

func (b *Body) AddReceiverTenantOutboundCommunicationIfMissing() error {
	if outboundCommunication := b.GetReceiverTenantOutboundCommunication(); !outboundCommunication.Exists() {
		if err := b.addReceiverTenantOutboundCommunication(); err != nil {
			return errors.Wrap(err, "while creating receiver tenant outbound communication")
		}
	}
	return nil
}

func (b *Body) addReceiverTenantOutboundCommunication() error {
	assignedTenantConfiguration := gjson.ParseBytes(b.AssignedTenant.Configuration)

	assignedTenantInboundCommunicationPath := FindKeyPath(assignedTenantConfiguration.Value(), inboundCommunicationKey)

	newConfiguration, err := sjson.SetBytes(b.ReceiverTenant.Configuration, strings.Replace(assignedTenantInboundCommunicationPath, inboundCommunicationKey, outboundCommunicationKey, 1), "{}")
	if err != nil {
		return err
	}
	b.ReceiverTenant.Configuration = newConfiguration

	return nil
}

func FindKeyPath(json interface{}, targetKey string) string {
	return findKeyPathHelper(json, targetKey, "")
}

func findKeyPathHelper(jsonData interface{}, targetKey string, currentPath string) string {
	switch v := jsonData.(type) {
	case map[string]interface{}:
		for key, _ := range v {
			if key == targetKey {
				return NewCurrentPath(currentPath, targetKey)
			}
		}

		for key, value := range v {
			if path := findKeyPathHelper(value, targetKey, NewCurrentPath(currentPath, key)); len(path) > 0 {
				return path
			}
		}
	case []interface{}:
		for i, item := range v {
			if path := findKeyPathHelper(item, targetKey, NewCurrentPath(currentPath, fmt.Sprint(i))); len(path) > 0 {
				return path
			}
		}
	}
	return ""
}

func NewCurrentPath(currentPath, key string) string {
	newPath := currentPath + "." + key
	if currentPath == "" {
		newPath = key
	}
	return newPath
}
