package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

var (
	ErrInvalidFormationID      = errors.New("$.formationId is invalid or missing")
	ErrInvalidAssignedTenantID = errors.New("$.assignedTenants[0].uclApplicationId is invalid or missing")
)

type TenantMapping struct {
	FormationID     string           `json:"formationId"`
	ReceiverTenant  ReceiverTenant   `json:"receiverTenant"`
	AssignedTenants []AssignedTenant `json:"assignedTenants"`
}

func (tm TenantMapping) String() string {
	if len(tm.AssignedTenants) == 0 {
		return fmt.Sprintf("$.formationId: %s, $.receiverTenant.applicationUrl: %s, no assigned tenants", tm.FormationID, tm.ReceiverTenant.ApplicationURL)
	}
	assignedTenant := tm.AssignedTenants[0]
	return fmt.Sprintf("$.formationId: '%s', $.receiverTenant.applicationUrl: '%s', $.assignedTenants[0]: (%s)",
		tm.FormationID, tm.ReceiverTenant.ApplicationURL, &assignedTenant)
}

type ReceiverTenant struct {
	ApplicationURL string `json:"applicationUrl"`
}

type (
	Operation       string
	State           string
	ApplicationType string
)

const (
	OperationAssign   Operation = "assign"
	OperationUnassign Operation = "unassign"

	StateInitial       State = "INITIAL"
	StateReady         State = "READY"
	StateConfigPending State = "CONFIG_PENDING"

	S4ApplicationType ApplicationType = "SAP S/4HANA Cloud"
)

type AssignedTenant struct {
	UCLApplicationID       string                      `json:"uclApplicationId"`
	UCLApplicationType     ApplicationType             `json:"uclApplicationType"`
	LocalTenantID          string                      `json:"localTenantId"`
	Operation              Operation                   `json:"operation"`
	ReverseAssignmentState State                       `json:"reverseAssignmentState"`
	Parameters             AssignedTenantParameters    `json:"parameters"`
	Config                 any                         `json:"configuration"`
	Configuration          AssignedTenantConfiguration `json:"-"`
}

func (at *AssignedTenant) String() string {
	return fmt.Sprintf(
		"$.operation: %s, $.localTenantId: %s, $.uclApplicationId: %s, $.uclApplicationType: %s, $.parameters.technicalIntegrationId: %s, $.configuration: %+v",
		at.Operation, at.LocalTenantID, at.UCLApplicationID, at.UCLApplicationType, at.Parameters.ClientID, at.Configuration)
}

func (at *AssignedTenant) SetConfiguration(ctx context.Context) error {
	log := logger.FromContext(ctx)

	if at.Config == nil {
		log.Info().Msg("$.assignedTenants[0].configuration is empty")
		return nil
	}
	b, err := json.Marshal(at.Config)
	if err != nil {
		return errors.Newf("failed to marshal $.assignedTenants[0].configuration: %w", err)
	}
	if err := json.Unmarshal(b, &at.Configuration); err != nil || len(at.Configuration.ConsumedAPIs) == 0 {
		log.Info().Msg("$.assignedTenants[0].configuration doesn't contain apis")
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
	if _, err := uuid.Parse(tm.AssignedTenants[0].UCLApplicationID); err != nil {
		return ErrInvalidAssignedTenantID
	}
	if tm.ReceiverTenant.ApplicationURL == "" {
		return errors.New("$.receiverTenant.applicationUrl is required")
	}
	if tm.AssignedTenants[0].LocalTenantID == "" {
		return errors.New("$.assignedTenants[0].localTenantId is required")
	}
	// TODO do we need to check the application type?
	if tm.AssignedTenants[0].UCLApplicationType == "" {
		return errors.New("$.assignedTenants[0].uclApplicationType is required")
	}
	if tm.AssignedTenants[0].Operation != OperationAssign && tm.AssignedTenants[0].Operation != OperationUnassign {
		return errors.New("$.assignedTenants[0].operation can only be assign or unassign")
	}
	if tm.AssignedTenants[0].Parameters.ClientID == "" {
		return errors.New("$.assignedTenants[0].parameters.technicalIntegrationId is required")
	}
	return nil
}

type TenantMappingResponse struct {
	State         State                      `json:"state"`
	Configuration TenantMappingConfiguration `json:"configuration"`
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
