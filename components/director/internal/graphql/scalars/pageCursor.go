package scalars

import (
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
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
