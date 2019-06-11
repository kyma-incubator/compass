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
		return errors.New("Error: can't convert input to string")
	}

	t, err := time.Parse(time.RFC3339, tmpStr)
	if err != nil {
		return errors.New("Error: can't parse time")
	}

	*y = Timestamp(t)

	return nil
}
func (y Timestamp) MarshalGQL(w io.Writer) {
	err := writeResponse(time.Time(y), w)
	if err != nil {
		log.Print(err)
		return
	}
}
