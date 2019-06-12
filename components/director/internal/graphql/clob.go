package graphql

import (
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
	"io"
	"log"
)

type CLOB []byte

func (y *CLOB) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToByteArray(v)
	if err != nil {
		return err
	}

	*y = CLOB(val)

	return nil
}

func (y CLOB) MarshalGQL(w io.Writer) {
	_, err := w.Write(y)
	if err != nil {
		log.Printf("error with writing %T", y)
	}
}
