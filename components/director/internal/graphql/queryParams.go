package graphql

import (
	"io"
	"log"
)

type QueryParams map[string][]string

func (y *QueryParams) UnmarshalGQL(v interface{}) error {
	params, err := convertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = params

	return nil
}

func (y QueryParams) MarshalGQL(w io.Writer) {
	err := marshalToWriter(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
