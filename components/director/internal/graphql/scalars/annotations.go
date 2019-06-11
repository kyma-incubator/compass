package scalars

import (
	"encoding/json"
	"io"
	"log"

	"github.com/pkg/errors"
)

type Annotations map[string]interface{}

func (y *Annotations) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return nil
	}

	val, ok := v.(string)
	if !ok {
		return errors.New("Error: can't convert input to byte array")
	}

	var result map[string]interface{}

	err := json.Unmarshal([]byte(val), &result)
	if err != nil {
		return errors.Errorf("Error with unmarshaling: %v", err)
	}

	*y = result

	return nil
}
func (y Annotations) MarshalGQL(w io.Writer) {
	err := writeResponse(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
