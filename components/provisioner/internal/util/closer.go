package util

import (
	"io"

	"github.com/sirupsen/logrus"
)

func Close(closer io.ReadCloser) {
	err := closer.Close()
	if err != nil {
		logrus.Warnf("Failed to close read closer: %s", err.Error())
	}
}
