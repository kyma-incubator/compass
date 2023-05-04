package types

import "encoding/json"

type Response struct {
	Configuration json.RawMessage `json:"configuration"`
}
