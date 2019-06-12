package scalars

import (
	"encoding/json"
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
	result := make(map[string]string)
	result["timestamp"] = time.Time(y).Format(time.RFC3339)

	bytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("error with marshalling %T", y)
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Error with writing %T", y)
	}
}
