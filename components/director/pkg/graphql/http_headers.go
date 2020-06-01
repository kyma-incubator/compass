package graphql

import (
	"encoding/json"
)

type HttpHeaders string

func (y *HttpHeaders) Unmarshal() (map[string][]string, error) {
	var data map[string][]string
	if y == nil {
		return data, nil
	}

	err := json.Unmarshal([]byte(*y), &data)

	return data, err
}

func NewHttpHeaders(h map[string][]string) (HttpHeaders, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return "", err
	}

	return HttpHeaders(data), nil
}
