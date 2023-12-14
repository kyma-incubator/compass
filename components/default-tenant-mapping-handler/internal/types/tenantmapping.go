package types

import (
	"fmt"
)

// TenantMapping represents the tenant mapping notification data
type TenantMapping struct {
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	AssignedTenant AssignedTenant `json:"assignedTenant"`
	Context        Context        `json:"context"`
}

// ReceiverTenant contains metadata for the receiver tenant - the tenant that receives the notification
type ReceiverTenant struct {
	State                string `json:"state"`
	AssignmentID         string `json:"uclAssignmentId"`
	DeploymentRegion     string `json:"deploymentRegion"`
	ApplicationNamespace string `json:"applicationNamespace"`
	ApplicationURL       string `json:"applicationUrl"`
	ApplicationTenantID  string `json:"applicationTenantId"`
	SubaccountID         string `json:"subaccountId"`
	Subdomain            string `json:"subdomain"`
	SystemName           string `json:"uclSystemName"`
	SystemTenantID       string `json:"uclSystemTenantId"`
}

// AssignedTenant contains metadata for the assigned tenant - the tenant that the notification is about
type AssignedTenant struct {
	State                string `json:"state"`
	AssignmentID         string `json:"uclAssignmentId"`
	DeploymentRegion     string `json:"deploymentRegion"`
	ApplicationNamespace string `json:"applicationNamespace"`
	ApplicationURL       string `json:"applicationUrl"`
	ApplicationTenantID  string `json:"applicationTenantId"`
	SubaccountID         string `json:"subaccountId"`
	Subdomain            string `json:"subdomain"`
	SystemName           string `json:"uclSystemName"`
	SystemTenantID       string `json:"uclSystemTenantId"`
}

// Context contains common data for the tenant mapping notification
type Context struct {
	Platform        string `json:"platform"`
	CrmID           string `json:"crmId"`
	AccountID       string `json:"accountId"`
	FormationID     string `json:"uclFormationId"`
	FormationName   string `json:"uclFormationName"`
	FormationTypeID string `json:"uclFormationTypeId"`
	Operation       string `json:"operation"`
}

// String prints the data in the tenant mapping notifications excepts the configuration part because it could have sensitive data.
// We should NOT include the configuration in this method.
func (tm *TenantMapping) String() string {
	return fmt.Sprintf("Context: {Platform: %s, CrmID: %s, AccountID: %s, FormationID: %s, FormationName: %s, FormationTypeID: %s, Operation: %s}, ReceiverTenant: {State: %s, AssignmentID: %s, DeploymentRegion: %s, ApplicationNamespace: %s, ApplicationURL: %s, ApplicationTenantID: %s, SubaccountID: %s, Subdomain: %s, SystemName: %s, SystemTenantID: %s}, AssignedTenant: {State: %s, AssignmentID: %s, DeploymentRegion: %s, ApplicationNamespace: %s, ApplicationURL: %s, ApplicationTenantID: %s, SubaccountID: %s, Subdomain: %s, SystemName: %s, SystemTenantID: %s}", tm.Context.Platform, tm.Context.CrmID, tm.Context.AccountID, tm.Context.FormationID, tm.Context.FormationName, tm.Context.FormationTypeID, tm.Context.Operation, tm.ReceiverTenant.State, tm.ReceiverTenant.AssignmentID, tm.ReceiverTenant.DeploymentRegion, tm.ReceiverTenant.ApplicationNamespace, tm.ReceiverTenant.ApplicationURL, tm.ReceiverTenant.ApplicationTenantID, tm.ReceiverTenant.SubaccountID, tm.ReceiverTenant.Subdomain, tm.ReceiverTenant.SystemName, tm.ReceiverTenant.SystemTenantID, tm.AssignedTenant.State, tm.AssignedTenant.AssignmentID, tm.AssignedTenant.DeploymentRegion, tm.AssignedTenant.ApplicationNamespace, tm.AssignedTenant.ApplicationURL, tm.AssignedTenant.ApplicationTenantID, tm.AssignedTenant.SubaccountID, tm.AssignedTenant.Subdomain, tm.AssignedTenant.SystemName, tm.AssignedTenant.SystemTenantID)
}
