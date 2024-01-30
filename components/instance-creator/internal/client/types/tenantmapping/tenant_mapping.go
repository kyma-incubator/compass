package tenantmapping

import (
	"encoding/json"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	assignOperation   = "assign"
	unassignOperation = "unassign"

	inboundCommunicationKey  = "inboundCommunication"
	outboundCommunicationKey = "outboundCommunication"
)

type TenantType int

const (
	AssignedTenantType TenantType = iota
	ReceiverTenantType
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

// GetTenantCommunication returns the Body tenant(Receiver or Assigned) communication(inbound or outbound)
func (b Body) GetTenantCommunication(tenantType TenantType, communicationType string) gjson.Result {
	var tenantConfiguration gjson.Result

	switch tenantType {
	case AssignedTenantType:
		tenantConfiguration = gjson.ParseBytes(b.AssignedTenant.Configuration)
	case ReceiverTenantType:
		tenantConfiguration = gjson.ParseBytes(b.ReceiverTenant.Configuration)
	default:
		return gjson.Result{}
	}

	communicationPath := FindKeyPath(tenantConfiguration.Value(), communicationType)
	if communicationPath == "" {
		return gjson.Result{}
	}

	return tenantConfiguration.Get(communicationPath)
}

// Validate validates the Body's Context
func (c Context) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.FormationID, validation.Required.Error("Context FormationID must be provided")),
		validation.Field(&c.Operation, validation.Required.Error("Context Operation must be provided"), validation.In(assignOperation, unassignOperation).Error(fmt.Sprintf("Context Operation must be %q or %q", assignOperation, unassignOperation))),
	)
}

// Validate validates the Body's ReceiverTenant
func (rt ReceiverTenant) Validate() error {
	return validation.ValidateStruct(&rt,
		validation.Field(&rt.Region, validation.Required.Error("ReceiverTenant Region must be provided")),
		validation.Field(&rt.SubaccountID, validation.Required.Error("ReceiverTenant SubaccountID must be provided")),
	)
}

// Validate validates the Body's AssignedTenant
func (at AssignedTenant) Validate() error {
	return validation.ValidateStruct(&at,
		validation.Field(&at.AssignmentID, validation.Required.Error("AssignedTenant AssignmentID must be provided")),
	)
}

// Validate validates the request Body
func (b Body) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.ReceiverTenant, validation.By(func(interface{}) error {
			return b.ReceiverTenant.Validate()
		})),
		validation.Field(&b.AssignedTenant, validation.By(func(interface{}) error {
			return b.AssignedTenant.Validate()
		})),
		validation.Field(&b.Context,
			validation.By(func(interface{}) error {
				return b.Context.Validate()
			}),
			validation.When(b.Context.Operation == assignOperation,
				validation.By(func(value interface{}) error {
					assignedTenantConfiguration := gjson.ParseBytes(b.AssignedTenant.Configuration)
					assignedTenantInboundCommunicationPath := FindKeyPath(assignedTenantConfiguration.Value(), inboundCommunicationKey)
					if assignedTenantInboundCommunicationPath == "" {
						return errors.New("AssignedTenant inbound communication is missing in the configuration")
					}

					receiverTenantConfiguration := gjson.ParseBytes(b.ReceiverTenant.Configuration)
					receiverTenantOutboundCommunicationPath := FindKeyPath(receiverTenantConfiguration.Value(), outboundCommunicationKey)
					if receiverTenantOutboundCommunicationPath != "" && strings.TrimSuffix(receiverTenantOutboundCommunicationPath, outboundCommunicationKey) != strings.TrimSuffix(assignedTenantInboundCommunicationPath, inboundCommunicationKey) {
						return errors.New("ReceiverTenant outbound communication should be in the same place as the assigned tenant inbound communication")
					}

					return nil
				}),
			)),
	)
}

func (b *Body) AddReceiverTenantOutboundCommunicationIfMissing() error {
	if outboundCommunication := b.GetTenantCommunication(ReceiverTenantType, outboundCommunicationKey); !outboundCommunication.Exists() {
		if err := b.addReceiverTenantOutboundCommunication(); err != nil {
			return errors.Wrap(err, "while creating receiver tenant outbound communication")
		}
	}
	return nil
}

func (b *Body) addReceiverTenantOutboundCommunication() error {
	assignedTenantConfiguration := gjson.ParseBytes(b.AssignedTenant.Configuration)

	assignedTenantInboundCommunicationPath := FindKeyPath(assignedTenantConfiguration.Value(), inboundCommunicationKey)

	newConfiguration, err := sjson.SetBytes(b.ReceiverTenant.Configuration, strings.Replace(assignedTenantInboundCommunicationPath, inboundCommunicationKey, outboundCommunicationKey, 1), map[string]string{})
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
		for key := range v {
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
	case string:
		if v == targetKey {
			return currentPath
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
