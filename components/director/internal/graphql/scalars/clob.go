package scalars

import (
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
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
