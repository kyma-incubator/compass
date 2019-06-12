package scalars

import (
	"io"
	"log"
	"time"

	"github.com/pkg/errors"
)

type Timestamp time.Time

func (y *Timestamp) UnmarshalGQL(v interface{}) error {

	if v == nil {
		return nil
	}

	tmpStr, ok := v.(string)

	if !ok {
		return errors.Errorf("unexpected input type: %T, should be string", v)

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
