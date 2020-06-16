package transformer_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/env"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/transformer"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	const testClientID = "testClientId"
	const testClientSecret = "testClientSecret"
	const testIssuerURL = "testIssuerURL"

	env.Config.OIDC.ClientID = testClientID
	env.Config.OIDC.ClientSecret = testClientSecret
	env.Config.OIDC.IssuerURL = testIssuerURL

	// Only pass t into top-level Convey calls
	Convey("NewClient()", t, func() {
		Convey("when given correct raw KubeConfig", func() {
			Convey("Should return a Client", func() {
				//given, when
				c, err := transformer.NewClient(testInputRawKubeconfig)
				//then
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

	Convey("client.TransformKubeconfig()", t, func() {
		Convey("Should return transformed kubeconfig", func() {
			//given
			c, err := transformer.NewClient(testInputRawKubeconfig)
			So(err, ShouldBeNil)
			//when
			res, err := c.TransformKubeconfig()
			//then
			So(err, ShouldBeNil)
			So(string(res), ShouldEqual, expectedTransformedKubeconfig)
		})
	})
}

var testInputRawKubeconfig = `
apiVersion: v1
kind: Config
clusters:
  - name: test--aa1234b
    cluster:
      server: 'https://api.kymatest.com'
      certificate-authority-data: LS0FakeFakeQo=
contexts:
  - name: test--aa1234b
    context:
      cluster: test--aa1234b
      user: test--aa1234b-token
current-context: test--aa1234b
users:
  - name: test--aa1234b-token
    user:
      token: 7WFakeFakeK
`

var expectedTransformedKubeconfig = `
---
apiVersion: v1
kind: Config
current-context: test--aa1234b
clusters:
- name: test--aa1234b
  cluster:
    certificate-authority-data: LS0FakeFakeQo=
    server: https://api.kymatest.com
contexts:
- name: test--aa1234b
  context:
    cluster: test--aa1234b
    user: test--aa1234b
users:
- name: test--aa1234b
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - "--oidc-issuer-url=testIssuerURL"
      - "--oidc-client-id=testClientId"
      - "--oidc-client-secret=testClientSecret"
      command: kubectl
`
