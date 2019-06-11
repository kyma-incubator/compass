package scalars

import (
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
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
