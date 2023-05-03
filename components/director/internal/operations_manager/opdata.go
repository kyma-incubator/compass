package operations_manager

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

func (b *OrdOperationData) getData() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling ord operation data")
	}

	return string(data), nil
}
