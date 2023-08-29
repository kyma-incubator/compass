package operationsmanager

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// OrdOperationData represents ord operation data.
type OrdOperationData struct {
	ApplicationID         string `json:"applicationID"`
	ApplicationTemplateID string `json:"applicationTemplateID,omitempty"`
}

// NewOrdOperationData creates new OrdOperationData.
func NewOrdOperationData(appID, appTemplateID string) *OrdOperationData {
	return &OrdOperationData{
		ApplicationID:         appID,
		ApplicationTemplateID: appTemplateID,
	}
}

// ParseOrdOperationData creates new OrdOperationData from byte array.
func ParseOrdOperationData(data []byte) (*OrdOperationData, error) {
	ordOperationData := &OrdOperationData{}
	if err := json.Unmarshal(data, ordOperationData); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling ord operation data")
	}
	return ordOperationData, nil
}

// GetData builds ord operation data
func (b *OrdOperationData) GetData() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling ord operation data")
	}

	return string(data), nil
}

// Equal returns true if current data content is the same as other's data content
func (b *OrdOperationData) Equal(other *OrdOperationData) bool {
	return b.ApplicationID == other.ApplicationID && b.ApplicationTemplateID == other.ApplicationTemplateID
}
