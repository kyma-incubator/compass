package graphql

import (
	"encoding/json"
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

type HTTPHeaders map[string][]string

func (y *HTTPHeaders) UnmarshalGQL(v interface{}) error {
	headers, err := scalar.ConvertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = headers

	return nil
}

func (y HTTPHeaders) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(y, w)
	if err != nil {
		log.D().Printf("while writing %T: %s", y, err)
		return
	}
}

type HTTPHeadersSerialized string

func (y *HTTPHeadersSerialized) Unmarshal() (map[string][]string, error) {
	var data map[string][]string
	if y == nil {
		return data, nil
	}

	err := json.Unmarshal([]byte(*y), &data)
	if err != nil {
		return nil, apperrors.NewInvalidDataError("unable to unmarshal HTTP headers: %s", err.Error())
	}

	return data, nil
}

func NewHTTPHeadersSerialized(h map[string][]string) (HTTPHeadersSerialized, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return "", apperrors.NewInvalidDataError("unable to marshal HTTP headers: %s", err.Error())
	}

	return HTTPHeadersSerialized(data), nil
}
