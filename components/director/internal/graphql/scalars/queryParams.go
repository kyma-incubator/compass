package scalars

import (
	"io"
	"log"
)

type QueryParams map[string][]string

func (y *QueryParams) UnmarshalGQL(v interface{}) error {

	params, err := convertToMapStringArrString(v)
	if err != nil {
		return err
	}

	*y = params

	return nil
}
func (y QueryParams) MarshalGQL(w io.Writer) {
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
