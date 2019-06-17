package graphql

import (
	"io"
	"log"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

type Labels map[string][]string

func (y *Labels) UnmarshalGQL(v interface{}) error {
	labels, err := scalar.ConvertToMapStringStringArray(v)
	if err != nil {
		return err
	}

	*y = labels

	return nil
}

func (y Labels) MarshalGQL(w io.Writer) {
	err := scalar.WriteMarshalled(y, w)
	if err != nil {
		log.Printf("while writing %T: %s", y, err)
		return
	}
}
