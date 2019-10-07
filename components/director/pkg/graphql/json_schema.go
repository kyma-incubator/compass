package graphql

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
	log "github.com/sirupsen/logrus"
)

type JSONSchema string

func (j *JSONSchema) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	*j = JSONSchema(val)
	return nil
}

func (j JSONSchema) MarshalGQL(w io.Writer) {
	_, err := io.WriteString(w, strconv.Quote(string(j)))
	if err != nil {
		log.Errorf("while writing %T: %s", j, err)
	}
}

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
