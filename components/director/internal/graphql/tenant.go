package graphql

import (
	"io"
	log "github.com/sirupsen/logrus"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
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
	_, err := io.WriteString(w, strconv.Quote(string(y)))
	if err != nil {
		log.Errorf("while writing %T: %s", y, err)
	}
}
