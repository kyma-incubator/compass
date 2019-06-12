package scalars

import (
	"encoding/json"
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
	bytes, err := json.Marshal(y)
	if err != nil {
		log.Print(err)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Print(err)
		return
	}
}
