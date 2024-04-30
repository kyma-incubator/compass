package data

import (
	"encoding/json"
	"github.com/pkg/errors"
)

// SystemFieldDiscoveryOperationData represents system field discovery operation data.
type SystemFieldDiscoveryOperationData struct {
	ApplicationID string `json:"applicationID"`
	TenantID      string `json:"tenantID"`
}

// NewSystemFieldDiscoveryOperationData creates new SystemFieldDiscoveryOperationData.
func NewSystemFieldDiscoveryOperationData(appID, tenantID string) *SystemFieldDiscoveryOperationData {
	return &SystemFieldDiscoveryOperationData{
		ApplicationID: appID,
		TenantID:      tenantID,
	}
}

// GetData builds system field discovery operation data
func (b *SystemFieldDiscoveryOperationData) GetData() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling system field discovery operation data")
	}

	return string(data), nil
}
