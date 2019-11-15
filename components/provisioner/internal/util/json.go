package util

import (
	"bytes"
	"encoding/json"
)

func DecodeJson(jsonConfig string, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader([]byte(jsonConfig)))
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func EncodeJson(state interface{}) (string, error) {
	bytesEncoded, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(bytesEncoded), nil
}
