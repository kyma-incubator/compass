package tests

import (
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testctx.Init()
	exitVal := m.Run()
	os.Exit(exitVal)
}