package scalars

import (
	"encoding/json"
	"io"
	"log"
)

type CLOB string

func (y *CLOB) UnmarshalGQL(v interface{}) error {
	val, err := convertToString(v)
	if err != nil {
		return err
	}
	*y = CLOB(val)

	return nil
}
func (y CLOB) MarshalGQL(w io.Writer) {
	result := make(map[string]string)
	result["CLOB"] = string(y)

	bytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("error with marshalling %T", y)
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("error with writing %T", y)
	}
}
