package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"
)

const surprise = "DpKtJ4j9jDq"

// Page missing godoc
type Page struct {
	StartCursor string
	EndCursor   string
	HasNextPage bool
}

// DecodeOffsetCursor missing godoc
func DecodeOffsetCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}

	decodedValue, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, errors.Wrap(err, "cursor is not correct")
	}

	realCursor := strings.TrimPrefix(string(decodedValue), surprise)

	offset, err := strconv.Atoi(realCursor)
	if err != nil {
		return 0, errors.Wrap(err, "cursor is not correct")
	}

	if offset < 0 {
		return 0, apperrors.NewInvalidDataError("cursor is not correct")
	}

	return offset, nil
}

// EncodeNextOffsetCursor missing godoc
func EncodeNextOffsetCursor(offset, pageSize int) string {
	nextPage := pageSize + offset

	cursor := surprise + strconv.Itoa(nextPage)

	return base64.StdEncoding.EncodeToString([]byte(cursor))
}

// ConvertOffsetLimitAndOrderedColumnToSQL missing godoc
func ConvertOffsetLimitAndOrderedColumnToSQL(pageSize, offset int, orderedColumn string) (string, error) {
	if orderedColumn == "" {
		return "", apperrors.NewInvalidDataError("to use pagination you must provide column to order by")
	}

	if pageSize < 1 {
		return "", apperrors.NewInvalidDataError("page size cannot be smaller than 1")
	}

	if offset < 0 {
		return "", apperrors.NewInvalidDataError("offset cannot be smaller than 0")
	}

	return fmt.Sprintf(`ORDER BY %s LIMIT %d OFFSET %d`, orderedColumn, pageSize, offset), nil
}
