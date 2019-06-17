package graphql

import (
	"io"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"

	"github.com/pkg/errors"
)

type Annotations map[string]interface{}

func (y *Annotations) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return errors.New("input should not be nil")
	}

	value, ok := v.(map[string]interface{})
	if !ok {
		return errors.Errorf("unexpected Annotations type: %T, should be map[string]interface{}", v)
	}

	*y = value

	return nil
}

func (y Annotations) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(y, w)
	if err != nil {
		log.Errorf("while writing %T: %s", y, err)
		return
	}
}
