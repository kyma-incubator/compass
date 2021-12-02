package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/pairing"

	director_http "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/stretchr/testify/require"
)

func TestGettingTokenWithMTLSWorks(t *testing.T) {
	reqData := pairing.RequestData{
		Application: graphql.Application{
			Name: conf.TestApplicationName,
			BaseEntity: &graphql.BaseEntity{
				ID: conf.TestApplicationID,
			},
		},
		Tenant:     conf.TestTenant,
		ClientUser: conf.TestClientUser,
	}
	jsonReqData, err := json.Marshal(reqData)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, conf.MTLSPairingAdapterURL, strings.NewReader(string(jsonReqData)))
	require.NoError(t, err)

	client := http.Client{
		Transport: director_http.NewServiceAccountTokenTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respParsed := struct {
		Token string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&respParsed)
	require.NoError(t, err)
	require.NotEmpty(t, respParsed.Token)
}
