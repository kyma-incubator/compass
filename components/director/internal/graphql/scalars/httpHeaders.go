package scalars

import (
	"io"
	"log"
)

type HttpHeaders map[string][]string

func (y *HttpHeaders) UnmarshalGQL(v interface{}) error {
	headers, err := convertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = headers

	return nil
}

func (y HttpHeaders) MarshalGQL(w io.Writer) {
	err := marshalAndWrite(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
