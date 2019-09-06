package graphql

import (
	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"
)

type JSON string

func (j *JSON) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	*j = JSON(val)
	return nil
}

func (j JSON) MarshalGQL(w io.Writer) {
	_, err := io.WriteString(w, strconv.Quote(string(j)))
	if err != nil {
		log.Errorf("while writing %T: %s", j, err)
	}
}
