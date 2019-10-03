package gqlschema

import (
	"io"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/scalar"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type AdditionalProperties map[string]interface{}

func (p *AdditionalProperties) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return errors.New("input should not be nil")
	}

	value, ok := v.(map[string]interface{})
	if !ok {
		return errors.Errorf("unexpected AdditionalProperties type: %T, should be map[string]interface{}", v)
	}

	*p = value

	return nil
}

func (p AdditionalProperties) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(p, w)
	if err != nil {
		log.Errorf("while writing %T: %s", p, err)
		return
	}
}
