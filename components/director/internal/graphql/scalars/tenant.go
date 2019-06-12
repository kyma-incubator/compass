package scalars

import (
	"encoding/json"
	"io"
	"log"
)

type Tenant string

func (y *Tenant) UnmarshalGQL(v interface{}) error {
	val, err := convertToString(v)
	if err != nil {
		return err
	}
	*y = Tenant(val)

	return nil
}

func (y Tenant) MarshalGQL(w io.Writer) {
	result := make(map[string]string)
	result["tenant"] = string(y)

	bytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("error with marshalling %T", y)
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("error with writing %T", y)
	}
}
