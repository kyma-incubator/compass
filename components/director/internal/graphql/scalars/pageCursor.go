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
	_, err := w.Write([]byte(y))
	if err != nil {
		log.Printf("Error with writing %T", y)
	}
}
