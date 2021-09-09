package graphql

import (
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"

	"github.com/pkg/errors"
)

// Labels missing godoc
type Labels map[string]interface{}

// UnmarshalGQL missing godoc
func (y *Labels) UnmarshalGQL(v interface{}) error {
	if v == nil {
		return apperrors.NewInternalError("input should not be nil")
	}

	value, ok := v.(map[string]interface{})
	if !ok {
		return errors.Errorf("unexpected Labels type: %T, should be map[string]interface{}", v)
	}

	*y = value

	return nil
}

// MarshalGQL missing godoc
func (y Labels) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(y, w)
	if err != nil {
		log.D().Errorf("while writing %T: %s", y, err)
		return
	}
}
