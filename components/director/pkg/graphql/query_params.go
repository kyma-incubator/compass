package graphql

import (
	"encoding/json"
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
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
		log.D().Errorf("while writing %T: %s", y, err)
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
	if err != nil {
		return nil, apperrors.NewInvalidDataError("unable to unmarshal query parameters: %s", err.Error())
	}

	return data, nil
}

func NewQueryParamsSerialized(h map[string][]string) (QueryParamsSerialized, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return "", apperrors.NewInvalidDataError("unable to marshal query parameters: %s", err.Error())
	}

	return QueryParamsSerialized(data), nil
}
