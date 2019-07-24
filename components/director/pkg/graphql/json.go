package graphql

import (
	"encoding/json"
	"github.com/99designs/gqlgen/graphql"
	"github.com/go-siris/siris/core/errors"
	log "github.com/sirupsen/logrus"
	"io"
)

type JSON interface{}

func MarshalJSON(v interface{}) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			log.Errorf("while writing %T: %s", v, err)
			return
		}
	})
}

func UnmarshalJSON(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, errors.New("input should not be nil")
	}
	return v, nil
}
