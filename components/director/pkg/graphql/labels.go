package graphql

import (
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"

	"github.com/pkg/errors"
)

type Labels map[string]interface{}

func (y *Labels) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return errors.New("input should not be nil")
	}

	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	var labels map[string]interface{}
	err = json.Unmarshal([]byte(val), &labels)
	if err != nil {
		return errors.New("Label input is not a valid JSON")
	}

	*y = labels

	return nil
}

func (y Labels) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalledAndStringified(y, w)
	if err != nil {
		log.Errorf("while writing %T: %s", y, err)
		return
	}
}
