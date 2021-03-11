package connector

import (
	"testing"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
)

type HydratorClient struct {
	*testkit.HydratorClient
}

func NewHydratorClient(validatorURL string) *HydratorClient {
	return &HydratorClient{
		HydratorClient: testkit.NewHydratorClient(validatorURL),
	}
}

func (vc *HydratorClient) ResolveCertificateData(t *testing.T, headers map[string][]string) oathkeeper.AuthenticationSession {
	return vc.ExecuteHydratorRequest(t, "/v1/certificate/data/resolve", headers)
}
