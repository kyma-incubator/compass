package types

import "encoding/json"

type Response struct { // todo::: consider removing it?
	Configuration json.RawMessage `json:"configuration"`
}
