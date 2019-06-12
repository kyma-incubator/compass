package scalars

import (
	"io"
	"log"
	"time"
)

type Timestamp time.Time

func (y *Timestamp) UnmarshalGQL(v interface{}) error {
	tmpStr,err := convertToString(v)
	if err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, tmpStr)
	if err != nil {
		return err
	}

	*y = Timestamp(t)

	return nil
}

func (y Timestamp) MarshalGQL(w io.Writer) {
	_, err := w.Write([]byte(time.Time(y).Format(time.RFC3339)))
	if err != nil {
		log.Printf("Error with writing %T", y)
	}
}
