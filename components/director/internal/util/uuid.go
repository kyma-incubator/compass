package util

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// UUIDToInt64 converts UUID to int64
func UUIDToInt64(identifier uuid.UUID) (int64, error) {
	return StringToInt64(identifier.String())
}

// StringToInt64 converts string to int64
func StringToInt64(input string) (int64, error) {
	if len(strings.TrimSpace(input)) == 0 {
		return int64(0), errors.New("input cannot be empty")
	}
	h := sha256.New()
	h.Write([]byte(input))
	return hashToInt64(fmt.Sprintf("%x", h.Sum(nil)))
}

func hashToInt64(hash string) (int64, error) {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return int64(0), err
	}
	intValue := new(big.Int).SetBytes(hashBytes)
	return intValue.Int64(), nil
}
