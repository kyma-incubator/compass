package graphql

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
)

type JSON string

func UnmarshalGQL(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, errors.New("input should not be nil")
	}
	return v, nil
}

func MarshalGQL(v interface{}) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			log.Errorf("while writing %T: %s", v, err)
			return
		}
	})
}
