package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/config"
)

var conf = &config.PairingAdapterConfig{}

func TestMain(m *testing.M) {
	config.ReadConfig(conf)

	exitVal := m.Run()
	os.Exit(exitVal)
}
