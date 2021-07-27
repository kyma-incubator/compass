package tests

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logrus.Info("Async Operations tests are not implemented")

	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}

func TestAsyncOperations(t *testing.T) {

}
