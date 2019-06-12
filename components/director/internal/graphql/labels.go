package graphql

import (
	"io"
	"log"
)

type Labels map[string][]string

func (y *Labels) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return nil
	}

	labels, err := convertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = labels

	return nil
}

func (y Labels) MarshalGQL(w io.Writer) {
	err := marshalToWriter(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
