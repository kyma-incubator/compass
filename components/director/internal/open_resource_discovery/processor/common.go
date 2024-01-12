package processor

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"strconv"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

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

func addFieldToLogger(ctx context.Context, fieldName, fieldValue string) context.Context {
	logger := log.LoggerFromContext(ctx)
	logger = logger.WithField(fieldName, fieldValue)
	return log.ContextWithLogger(ctx, logger)
}

func checkIfShouldFetchSpecs(lastUpdateValueFromDoc, lastUpdateValueFromDB *string) (bool, error) {
	if lastUpdateValueFromDoc == nil || lastUpdateValueFromDB == nil {
		return true, nil
	}

	lastUpdateTimeFromDoc, err := time.Parse(time.RFC3339, str.PtrStrToStr(lastUpdateValueFromDoc))
	if err != nil {
		return false, err
	}

	lastUpdateTimeFromDB, err := time.Parse(time.RFC3339, str.PtrStrToStr(lastUpdateValueFromDB))
	if err != nil {
		return false, err
	}

	return lastUpdateTimeFromDoc.After(lastUpdateTimeFromDB), nil
}

func NewestLastUpdateTimestamp(lastUpdateValueFromDoc, lastUpdateValueFromDB, hashFromDB *string, hashFromDoc uint64) (*string, error) {
	newestLastUpdateTime, err := compareLastUpdateFromDocAndDB(lastUpdateValueFromDoc, lastUpdateValueFromDB)
	if err != nil {
		return nil, err
	}

	var hashIsEqual bool
	if hashFromDB != nil {
		hashIsEqual = cmp.Equal(*hashFromDB, strconv.FormatUint(hashFromDoc, 10))
	}

	if !hashIsEqual {
		currentTime := time.Now().Format(time.RFC3339)
		newestLastUpdateTime = &currentTime
	}

	return newestLastUpdateTime, nil
}

func compareLastUpdateFromDocAndDB(lastUpdateValueFromDoc, lastUpdateValueFromDB *string) (*string, error) {
	if lastUpdateValueFromDoc == nil {
		if lastUpdateValueFromDB == nil {
			currentTime := time.Now().Format(time.RFC3339)
			return &currentTime, nil
		}
		return lastUpdateValueFromDB, nil
	}

	if lastUpdateValueFromDB == nil {
		return lastUpdateValueFromDoc, nil
	}

	newestLastUpdateTime := lastUpdateValueFromDB

	lastUpdateTimeFromDoc, err := time.Parse(time.RFC3339, str.PtrStrToStr(lastUpdateValueFromDoc))
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing lastUpdate timestamp from document")
	}

	lastUpdateTimeFromDB, err := time.Parse(time.RFC3339, str.PtrStrToStr(lastUpdateValueFromDB))
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing lastUpdate timestamp from db")
	}

	if lastUpdateTimeFromDoc.After(lastUpdateTimeFromDB) {
		newestLastUpdateTime = lastUpdateValueFromDoc
	}

	return newestLastUpdateTime, nil
}
