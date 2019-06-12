package scalars

import (
	"io"
	"log"

	"github.com/pkg/errors"
)

type Annotations map[string]interface{}

func (y *Annotations) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return nil
	}
	value, ok := v.(map[string]interface{})
	if !ok {
		return errors.Errorf("unexpected Annotations type: %T, should be map[string]interface{}", v)
	}

	*y = value

	return nil
}

func (y Annotations) MarshalGQL(w io.Writer) {
	err := marshalAndWrite(y, w)
	if err != nil {
		log.Print(err)
		return
	}
}
