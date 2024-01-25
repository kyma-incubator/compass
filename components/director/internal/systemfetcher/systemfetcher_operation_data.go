package systemfetcher

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// SystemFetcherOperationData represents system fetcher operation data.
type SystemFetcherOperationData struct {
	TenantID string `json:"tenantID"`
}

// NewSystemFetcherOperationData creates new SystemFetcherOperationData.
func NewSystemFetcherOperationData(tenantID string) *SystemFetcherOperationData {
	return &SystemFetcherOperationData{
		TenantID: tenantID,
	}
}

// ParseSystemFetcherOperationData creates new SystemFetcherOperationData from byte array.
func ParseSystemFetcherOperationData(data []byte) (*SystemFetcherOperationData, error) {
	systemFetcherOperationData := &SystemFetcherOperationData{}
	if err := json.Unmarshal(data, systemFetcherOperationData); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling system fetcher operation data")
	}
	return systemFetcherOperationData, nil
}

// GetData builds system fetcher operation data
func (b *SystemFetcherOperationData) GetData() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling system fetcher operation data")
	}

	return string(data), nil
}

// Equal returns true if current data content is the same as other's data content
func (b *SystemFetcherOperationData) Equal(other *SystemFetcherOperationData) bool {
	return b.TenantID == other.TenantID
}
