package graphql

import (
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
	"io"
	"log"
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
		log.Print(err)
		return
	}
}
