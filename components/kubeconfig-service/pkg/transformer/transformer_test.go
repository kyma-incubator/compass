package transformer

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/env"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	const testClientID = "testClientId"
	const testClientSecret = "testClientSecret"
	const testIssuerURL = "testIssuerURL"

	// Only pass t into top-level Convey calls
	Convey("NewClient()", t, func() {
		Convey("when given correct raw KubeConfig", func() {
			env.Config.OIDC.ClientID = testClientID
			env.Config.OIDC.ClientSecret = testClientSecret
			env.Config.OIDC.IssuerURL = testIssuerURL

			Convey("Should return a Client", func() {
				c, err := NewClient(rawKubeCfg)
				So(err, ShouldBeNil)
				So(c.ContextName, ShouldEqual, "test--aa1234b")
				So(c.CAData, ShouldEqual, "LS0FakeFakeQo=")
				So(c.ServerURL, ShouldEqual, "https://api.kymatest.com")
				So(c.OIDCClientID, ShouldEqual, testClientID)
				So(c.OIDCClientSecret, ShouldEqual, testClientSecret)
				So(c.OIDCIssuerURL, ShouldEqual, testIssuerURL)
			})
		})
	})
}

var rawKubeCfg = `
apiVersion: v1
kind: Config
clusters:
  - name: test--aa1234b
    cluster:
      server: 'https://api.kymatest.com'
      certificate-authority-data: >-
        LS0FakeFakeQo=
contexts:
  - name: test--aa1234b
    context:
      cluster: test--aa1234b
      user: test--aa1234b-token
current-context: test--aa1234b
users:
  - name: test--aa1234b-token
    user:
      token: >-
        7WFakeFakeK
`
