package graphql

import (
	"io"
	"log"

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
