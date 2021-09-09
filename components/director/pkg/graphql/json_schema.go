package graphql

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

// JSONSchema missing godoc
type JSONSchema string

// UnmarshalGQL missing godoc
func (j *JSONSchema) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(val), new(interface{}))
	if err != nil {
		return apperrors.NewInvalidDataError("JSONSchema input is not a valid JSON")
	}

	*j = JSONSchema(val)
	return nil
}

// MarshalGQL missing godoc
func (j JSONSchema) MarshalGQL(w io.Writer) {
	_, err := io.WriteString(w, strconv.Quote(string(j)))
	if err != nil {
		log.D().Errorf("while writing %T: %s", j, err)
	}
}

// Unmarshal missing godoc
func (j *JSONSchema) Unmarshal() (*interface{}, error) {
	if j == nil {
		return nil, nil
	}
	var output interface{}
	err := json.Unmarshal([]byte(*j), &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

// MarshalSchema missing godoc
func MarshalSchema(input *interface{}) (*JSONSchema, error) {
	if input == nil {
		return nil, nil
	}
	schema, err := json.Marshal(*input)
	if err != nil {
		return nil, err
	}
	output := JSONSchema(string(schema))
	return &output, nil
}
