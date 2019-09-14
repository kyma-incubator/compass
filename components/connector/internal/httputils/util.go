package httputils

import (
	"io"

	"github.com/sirupsen/logrus"
)

func Close(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		logrus.Warnf("Warning: failed to close: %s", err.Error())
	}
}
