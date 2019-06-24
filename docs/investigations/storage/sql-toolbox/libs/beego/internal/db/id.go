package db

import "github.com/satori/go.uuid"

type IDProvider struct {
}

func (i *IDProvider) GenID() (string, error) {
	if u, err := uuid.NewV4(); err != nil {
		return "", err
	} else {

		return u.String(), nil
	}
}
