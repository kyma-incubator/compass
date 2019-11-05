package util

import (
	"bytes"
	"encoding/json"
)

func DecodeJson(providerSpecificConfig string, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader([]byte(providerSpecificConfig)))
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}
