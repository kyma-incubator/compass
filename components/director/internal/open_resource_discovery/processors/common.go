package processors

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/mitchellh/hashstructure/v2"
)

func addFieldToLogger(ctx context.Context, fieldName, fieldValue string) context.Context {
	logger := log.LoggerFromContext(ctx)
	logger = logger.WithField(fieldName, fieldValue)
	return log.ContextWithLogger(ctx, logger)
}

func searchInSlice(length int, f func(i int) bool) (int, bool) {
	for i := 0; i < length; i++ {
		if f(i) {
			return i, true
		}
	}
	return -1, false
}

func equalStrings(first, second *string) bool {
	return first != nil && second != nil && *first == *second
}

// HashObject hashes the given object
func HashObject(obj interface{}) (uint64, error) {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		return 0, errors.New("failed to hash the given object")
	}

	return hash, nil
}
