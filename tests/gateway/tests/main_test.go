package tests

import (
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"log"
	"os"
	"testing"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	DefaultTenant string
	Domain        string
	DirectorURL   string
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}
	testctx.Init()
	testConfig.DirectorURL = fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", testConfig.Domain)
	exitVal := m.Run()
	os.Exit(exitVal)

}
