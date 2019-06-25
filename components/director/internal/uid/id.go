package uid

import "github.com/google/uuid"

func Generate() string {
	return uuid.New().String()
}
