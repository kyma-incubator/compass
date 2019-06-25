package graphql

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

type CLOB string

func (y *CLOB) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	*y = CLOB(val)

	return nil
}

func (y CLOB) MarshalGQL(w io.Writer) {
	json := fmt.Sprintf(`"%s"`,y)
	_, err := w.Write([]byte(json))
	if err != nil {
		log.Errorf("while writing %T: %s", y, err)
	}
}
