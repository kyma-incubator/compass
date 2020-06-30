package graphql

import (
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

type QueryParams map[string][]string

func (y *QueryParams) UnmarshalGQL(v interface{}) error {
	params, err := scalar.ConvertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = params

	return nil
}

func (y QueryParams) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(y, w)
	if err != nil {
		log.Errorf("while writing %T: %s", y, err)
		return
	}
}

type QueryParamsSerialized string

func (y *QueryParamsSerialized) Unmarshal() (map[string][]string, error) {
	var data map[string][]string
	if y == nil {
		return data, nil
	}

	err := json.Unmarshal([]byte(*y), &data)

	return data, err
}

func NewQueryParamsSerialized(h map[string][]string) (QueryParamsSerialized, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return "", err
	}

	return QueryParamsSerialized(data), nil
}
