package pagination

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

const surprise = "DpKtJ4j9jDq"

type Page struct {
	StartCursor string
	EndCursor   string
	HasNextPage bool
}

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
		return 0, errors.New("cursor is not correct")
	}

	return offset, nil
}

func EncodeOffsetCursor(offset, pageSize int) string {
	nextPage := pageSize + offset

	cursor := surprise + strconv.Itoa(nextPage)

	return base64.StdEncoding.EncodeToString([]byte(cursor))
}

func ConvertOffsetLimitAndOrderedColumnToSQL(pageSize, offset int, orderedColumn string) (string, error) {
	if orderedColumn == "" {
		return "", errors.New("to use pagination you must provide column to order by")
	}
	return fmt.Sprintf(` ORDER BY "%s" LIMIT %d OFFSET %d`, orderedColumn, pageSize, offset), nil
}
