package graphql

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"
)

type JSON string

func (j *JSON) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	*j = JSON(val)
	return nil
}

func (j JSON) MarshalGQL(w io.Writer) {
	_, err := io.WriteString(w, strconv.Quote(string(j)))
	if err != nil {
		log.Errorf("while writing %T: %s", j, err)
	}
}

func (j *JSON) UnmarshalSchema() (*interface{}, error) {
	if j == nil {
		return nil, nil;
	}
	var output interface{}
	err := json.Unmarshal([]byte(*j), &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func MarshalSchema(input *interface{}) (*JSON, error) {
	if input == nil {
		return nil, nil
	}
	schema, err := json.Marshal(*input)
	if err != nil {
		return nil, err
	}
	output := JSON(string(schema))
	return &output, nil
}
