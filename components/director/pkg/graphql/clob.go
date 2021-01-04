package graphql

import (
	"io"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

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
	_, err := io.WriteString(w, strconv.Quote(string(y)))
	if err != nil {
		log.D().Errorf("while writing %T: %s", y, err)
	}
}
