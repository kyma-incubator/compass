package graphql

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

// JSON missing godoc
type JSON string

// UnmarshalGQL missing godoc
func (j *JSON) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(val), new(interface{}))
	if err != nil {
		return apperrors.NewInternalError("JSON input is not a valid JSON")
	}

	*j = JSON(val)
	return nil
}

// MarshalGQL missing godoc
func (j JSON) MarshalGQL(w io.Writer) {
	_, err := io.WriteString(w, strconv.Quote(string(j)))
	if err != nil {
		log.D().Errorf("while writing %T: %s", j, err)
	}
}
