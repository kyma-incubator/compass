package tenantmapping

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
func (b Body) GetAssignedTenantInboundCommunication() (map[string]json.RawMessage, error) {
	var assignedTenantConfiguration interface{}
	if err := json.Unmarshal(b.AssignedTenant.Configuration, &assignedTenantConfiguration); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling assigned tenant configuration")
	}

	assignedTenantInboundCommunicationPath := findKeyPath(assignedTenantConfiguration, "inboundCommunication", "")
	if assignedTenantInboundCommunicationPath == "" {
		return nil, errors.New("Assigned tenant inbound communication is missing in the configuration")
	}

	assignedTenantInboundCommunication := gjson.GetBytes(b.AssignedTenant.Configuration, assignedTenantInboundCommunicationPath+".inboundCommunication")
	if !assignedTenantInboundCommunication.Exists() {
		return nil, errors.New("Assigned tenant inbound communication is missing in the configuration")
	}

	marshalledInboundCommunication, err := json.Marshal(assignedTenantInboundCommunication.Value())
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling assigned tenant inbound communication")
	}

	var res map[string]json.RawMessage
	if err := json.Unmarshal(marshalledInboundCommunication, &res); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling assigned tenant inbound communication to map of string to json.RawMessage")
	}

	return res, nil
}

// GetReceiverTenantOutboundCommunication returns the Body receiver tenant outbound communication
func (b Body) GetReceiverTenantOutboundCommunication() (map[string]json.RawMessage, error) {
	var receiverTenantConfiguration interface{}
	if err := json.Unmarshal(b.ReceiverTenant.Configuration, &receiverTenantConfiguration); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling receiver tenant configuration")
	}

	receiverTenantInboundCommunicationPath := findKeyPath(receiverTenantConfiguration, "outboundCommunication", "")
	if receiverTenantInboundCommunicationPath == "" {
		return nil, errors.New("Receiver tenant outbound communication is missing in the configuration")
	}

	receiverTenantInboundCommunication := gjson.GetBytes(b.ReceiverTenant.Configuration, receiverTenantInboundCommunicationPath+".outboundCommunication")
	if !receiverTenantInboundCommunication.Exists() {
		return nil, errors.New("Receiver tenant outbound communication is missing in the configuration")
	}

	marshalledOutboundCommunication, err := json.Marshal(receiverTenantInboundCommunication.Value())
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling receiver tenant outbound communication")
	}

	var res map[string]json.RawMessage
	if err := json.Unmarshal(marshalledOutboundCommunication, &res); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling receiver tenant outbound communication to map of string to json.RawMessage")
	}

	return res, nil
}

func (b *Body) SetReceiverTenantAuth(authKey string, authValue map[string]interface{}) error {
	var receiverTenantConfiguration interface{}
	if err := json.Unmarshal(b.ReceiverTenant.Configuration, &receiverTenantConfiguration); err != nil {
		return errors.Wrap(err, "while unmarshalling receiver tenant outbound communication")
	}

	receiverTenantInboundCommunicationPath := findKeyPath(receiverTenantConfiguration, "outboundCommunication", "")
	if receiverTenantInboundCommunicationPath == "" {
		return errors.New("Receiver tenant inbound communication is missing in the configuration")
	}

	newReceiverTenantConfiguration, err := sjson.SetBytes(b.ReceiverTenant.Configuration, fmt.Sprintf("%s.outboundCommunication.%s", receiverTenantInboundCommunicationPath, authKey), authValue)
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

	if b.Context.Operation == "" || (b.Context.Operation != "assign" && b.Context.Operation != "unassign") {
		return apperrors.NewInvalidDataError("Context's Operation is invalid, expected %q or %q, got: %q", "assign", "unassign", b.Context.Operation)
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

	if b.Context.Operation == "assign" {
		var assignedTenantConfiguration interface{}
		if err := json.Unmarshal(b.AssignedTenant.Configuration, &assignedTenantConfiguration); err != nil {
			return apperrors.NewInvalidDataError("While unmarshalling assigned tenant configuration")
		}

		assignedTenantInboundCommunicationPath := findKeyPath(assignedTenantConfiguration, "inboundCommunication", "")
		if assignedTenantInboundCommunicationPath == "" {
			return apperrors.NewInvalidDataError("Assigned tenant inbound communication is missing in the configuration")
		}

		if gjson.GetBytes(b.ReceiverTenant.Configuration, assignedTenantInboundCommunicationPath+".outboundCommunication").Exists() == false {
			return apperrors.NewInvalidDataError(fmt.Sprintf("Receiver tenant outbound communication is missing - should be in %q in the configuration", assignedTenantInboundCommunicationPath))
		}
	}

	return nil
}

func findKeyPath(json interface{}, targetKey string, currentPath string) string {
	switch v := json.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == targetKey {
				return currentPath
			}
			if path := findKeyPath(value, targetKey, newCurrentPath(currentPath, key)); len(path) > 0 {
				return path
			}
		}
	case []interface{}:
		for i, item := range v {
			if path := findKeyPath(item, targetKey, newCurrentPath(currentPath, fmt.Sprint(i))); len(path) > 0 {
				return path
			}
		}
	}
	return ""
}

func newCurrentPath(currentPath, key string) string {
	newPath := currentPath + "." + key
	if currentPath == "" {
		newPath = key
	}
	return newPath
}
