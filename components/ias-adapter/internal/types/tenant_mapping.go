package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

const (
	S4SAPManagedCommunicationScenario = "SAP_COM_1002"
)

var (
	ErrInvalidFormationID         = errors.New("$.context.uclFormationId is invalid or missing")
	ErrInvalidAssignedTenantAppID = errors.New("$.assignedTenant.uclSystemTenantId is invalid or missing")
)

type TenantMapping struct {
	Context        `json:"context"`
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	AssignedTenant AssignedTenant `json:"assignedTenant"`
}

type Context struct {
	FormationID string    `json:"uclFormationId"`
	Operation   Operation `json:"operation"`
}

func (tm TenantMapping) String() string {
	return fmt.Sprintf("$.context.uclFormationId: '%s', $.context.operation: '%s', $.receiverTenant.applicationUrl: '%s', $.assignedTenant: (%s)",
		tm.FormationID, tm.Operation, tm.ReceiverTenant.ApplicationURL, tm.AssignedTenant)
}

type ReceiverTenant struct {
	ApplicationURL string `json:"applicationUrl"`
}

type (
	Operation            string
	State                string
	ApplicationNamespace string
)

const (
	OperationAssign   Operation = "assign"
	OperationUnassign Operation = "unassign"

	StateInitial       State = "INITIAL"
	StateConfigPending State = "CONFIG_PENDING"
	StateCreateError   State = "CREATE_ERROR"
	StateDeleteError   State = "DELETE_ERROR"
	StateCreateReady   State = "CREATE_READY"
	StateDeleteReady   State = "DELETE_READY"
	StateReady         State = "READY"

	S4ApplicationNamespace ApplicationNamespace = "sap.s4"
)

func ReadyState(operation Operation) State {
	if operation == OperationAssign {
		return StateCreateReady
	}
	return StateDeleteReady
}

func ErrorState(operation Operation) State {
	if operation == OperationAssign {
		return StateCreateError
	}
	return StateDeleteError
}

type AssignedTenant struct {
	AppID                  string                      `json:"uclSystemTenantId"`
	AppNamespace           ApplicationNamespace        `json:"applicationNamespace"`
	LocalTenantID          string                      `json:"applicationTenantId"`
	ReverseAssignmentState State                       `json:"state"`
	Parameters             AssignedTenantParameters    `json:"parameters"`
	Config                 any                         `json:"configuration"`
	Configuration          AssignedTenantConfiguration `json:"-"`
}

func (at *AssignedTenant) String() string {
	return fmt.Sprintf(
		"$.applicationTenantId: %s, $.uclSystemTenantId: %s, $.applicationNamespace: %s, $.parameters.technicalIntegrationId: %s, $.configuration: %+v",
		at.LocalTenantID, at.AppID, at.AppNamespace, at.Parameters.ClientID, at.Configuration)
}

func (at *AssignedTenant) SetConfiguration(ctx context.Context) error {
	log := logger.FromContext(ctx)

	if at.Config == nil {
		log.Info().Msg("$.assignedTenant.configuration is empty")
		return nil
	}
	b, err := json.Marshal(at.Config)
	if err != nil {
		return errors.Newf("failed to marshal $.assignedTenant.configuration: %w", err)
	}
	if err := json.Unmarshal(b, &at.Configuration); err != nil || len(at.Configuration.ConsumedAPIs) == 0 {
		log.Info().Msg("$.assignedTenant.configuration doesn't contain apis")
		return nil
	}

	return nil
}

type AssignedTenantParameters struct {
	ClientID         string `json:"technicalIntegrationId"`
	IASApplicationID string
}

type AssignedTenantConfiguration struct {
	ConsumedAPIs []string    `json:"apis"`
	Credentials  Credentials `json:"credentials"`
}

func (tm TenantMapping) Validate() error {
	if _, err := uuid.Parse(tm.FormationID); err != nil {
		return ErrInvalidFormationID
	}
	if tm.Operation != OperationAssign && tm.Operation != OperationUnassign {
		return errors.New("$.context.operation can only be assign or unassign")
	}
	if _, err := uuid.Parse(tm.AssignedTenant.AppID); err != nil {
		return ErrInvalidAssignedTenantAppID
	}
	if tm.ReceiverTenant.ApplicationURL == "" {
		return errors.New("$.receiverTenant.applicationUrl is required")
	}
	if tm.AssignedTenant.LocalTenantID == "" {
		return errors.New("$.assignedTenant.applicationTenantId is required")
	}
	if tm.AssignedTenant.AppNamespace == "" {
		return errors.New("$.assignedTenant.applicationNamespace is required")
	}
	// S/4 applications are created by the IAS adapter and therefore the tenant mapping does not contain its clientID
	if tm.AssignedTenant.AppNamespace != S4ApplicationNamespace && tm.AssignedTenant.Parameters.ClientID == "" {
		return errors.New("$.assignedTenant.parameters.technicalIntegrationId is required")
	}
	return nil
}

type TenantMappingConfiguration struct {
	Credentials Credentials `json:"credentials"`
}

type Credentials struct {
	OutboundCommunicationCredentials CommunicationCredentials `json:"outboundCommunication"`
	InboundCommunicationCredentials  CommunicationCredentials `json:"inboundCommunication"`
}

type CommunicationCredentials struct {
	OAuth2mTLSAuthentication OAuth2mTLSAuthentication `json:"oauth2mtls"`
}

type OAuth2mTLSAuthentication struct {
	CorrelationIds []string `json:"correlationIds"`
	Certificate    string   `json:"certificate"`
}
