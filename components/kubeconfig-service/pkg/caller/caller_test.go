package caller_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/caller"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testTenant      = "testTenant1"
	testRuntimeID   = "testRuntime1"
	mockGQLResponse = `{
		"data": {
			"result": {
				"runtimeConfiguration": {
					"kubeconfig": "%s"
				}
			}
		}
	}`
)

func TestSpec(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("Caller", t, func() {
		Convey("RuntimeStatus()", func() {
			Convey("Should return a RuntimeStatus in a happy path scenario", func(c C) {

				//given
				testKubeconfig := "some-test-kubeconfig"

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					//Assertion on tenant parameter passed in a HTTP header
					c.So(r.Header.Get(caller.TenantHeader), ShouldEqual, testTenant)

					b, err := ioutil.ReadAll(r.Body)
					c.So(err, ShouldBeNil)

					//Assertion on runtimeStatus parameter (embedded in the query)
					c.So(string(b), ShouldContainSubstring, fmt.Sprintf(`result: runtimeStatus(id: \"%s\")`, testRuntimeID))

					//Mock response, our contract on what the endpoint shall return
					_, err = io.WriteString(w, fmt.Sprintf(mockGQLResponse, testKubeconfig))
					c.So(err, ShouldBeNil)
				}))
				defer srv.Close()

				cllr := caller.NewCaller(srv.URL, testTenant)

				//when
				res, err := cllr.RuntimeStatus(testRuntimeID)

				//then
				So(err, ShouldBeNil)
				So(*res.RuntimeConfiguration.Kubeconfig, ShouldEqual, testKubeconfig)
			})
		})
	})
}
