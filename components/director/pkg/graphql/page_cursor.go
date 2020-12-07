package graphql

import (
	"io"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/scalar"
)

type PageCursor string

func (y *PageCursor) UnmarshalGQL(v interface{}) error {
	val, err := scalar.ConvertToString(v)
	if err != nil {
		return err
	}

	*y = PageCursor(val)

	return nil
}

func (y PageCursor) MarshalGQL(w io.Writer) {
	_, err := io.WriteString(w, strconv.Quote(string(y)))
	if err != nil {
		log.D().Errorf("while writing %T: %s", y, err)
	}
}
