package graphql

import (
	"encoding/json"
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

// QueryParams missing godoc
type QueryParams map[string][]string

// UnmarshalGQL missing godoc
func (y *QueryParams) UnmarshalGQL(v interface{}) error {
	params, err := scalar.ConvertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = params

	return nil
}

// MarshalGQL missing godoc
func (y QueryParams) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(y, w)
	if err != nil {
		log.D().Errorf("while writing %T: %s", y, err)
		return
	}
}

// QueryParamsSerialized missing godoc
type QueryParamsSerialized string

// Unmarshal missing godoc
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

// NewQueryParamsSerialized missing godoc
func NewQueryParamsSerialized(h map[string][]string) (QueryParamsSerialized, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return "", apperrors.NewInvalidDataError("unable to marshal query parameters: %s", err.Error())
	}

	return QueryParamsSerialized(data), nil
}
