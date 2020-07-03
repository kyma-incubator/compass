package graphql

import (
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

type HttpHeaders map[string][]string

func (y *HttpHeaders) UnmarshalGQL(v interface{}) error {
	headers, err := scalar.ConvertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = headers

	return nil
}

func (y HttpHeaders) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(y, w)
	if err != nil {
		log.Printf("while writing %T: %s", y, err)
		return
	}
}

type HttpHeadersSerialized string

func (y *HttpHeadersSerialized) Unmarshal() (map[string][]string, error) {
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

func NewHttpHeadersSerialized(h map[string][]string) (HttpHeadersSerialized, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return "", apperrors.NewInvalidDataError("unable to marshal HTTP headers: %s", err.Error())
	}

	return HttpHeadersSerialized(data), nil
}
