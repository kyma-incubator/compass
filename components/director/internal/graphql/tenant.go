package graphql

import (
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
	"io"
	"log"
)

type Tenant string

func (y *Tenant) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	*y = Tenant(val)

	return nil
}

func (y Tenant) MarshalGQL(w io.Writer) {
	_, err := w.Write([]byte(y))
	if err != nil {
		log.Printf("error with writing %T", y)
	}
}
