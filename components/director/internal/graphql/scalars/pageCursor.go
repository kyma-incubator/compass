package scalars

import (
	"encoding/json"
	"io"
	"log"
)

type PageCursor string

func (y *PageCursor) UnmarshalGQL(v interface{}) error {
	val, err := convertToString(v)
	if err != nil {
		return err
	}
	*y = PageCursor(val)

	return nil
}
func (y PageCursor) MarshalGQL(w io.Writer) {
	result := make(map[string]string)
	result["pageCursor"] = string(y)

	bytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("error with marshalling %T", y)
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Error with writing %T", y)
	}
}
