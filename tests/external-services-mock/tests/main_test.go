package tests

import (
	"log"
	"os"
	"testing"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	DefaultTenant               string
	DirectorURL                 string
	ExternalServicesMockBaseURL string
	BasicCredentialsUsername    string
	BasicCredentialsPassword    string
	AppClientID                 string
	AppClientSecret             string
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}
