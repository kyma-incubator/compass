package graphql

import (
	"io"
	"strconv"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

// Timestamp missing godoc
type Timestamp time.Time

// UnmarshalGQL missing godoc
func (y *Timestamp) UnmarshalGQL(v interface{}) error {
	tmpStr, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, tmpStr)
	if err != nil {
		return err
	}

	*y = Timestamp(t)

	return nil
}

// MarshalGQL missing godoc
func (y Timestamp) MarshalGQL(w io.Writer) {
	_, err := w.Write([]byte(strconv.Quote(time.Time(y).Format(time.RFC3339))))
	if err != nil {
		log.D().Errorf("while writing %T: %s", y, err)
	}
}

// MarshalJSON missing godoc
func (y Timestamp) MarshalJSON() ([]byte, error) {
	return time.Time(y).MarshalJSON()
}

// UnmarshalJSON missing godoc
func (y *Timestamp) UnmarshalJSON(data []byte) error {
	return (*time.Time)(y).UnmarshalJSON(data)
}
