package ias

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIASService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IAS Service Test Suite")
}
