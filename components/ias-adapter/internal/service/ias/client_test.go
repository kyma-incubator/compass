package ias

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/middlewares"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IAS Client Test Suite")
}

var _ = Describe("IAS Client", func() {
	When("Outbound request is sent", func() {
		It("Correlation ID header should be present", func() {
			testClient := &http.Client{
				Transport: &headerTransport{clientTransport: http.DefaultTransport},
			}
			correlationID := "123"
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Header.Get(middlewares.CorrelationIDHeader)).To(Equal(correlationID))
			}))
			defer testServer.Close()

			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Set(middlewares.CorrelationIDKey, correlationID)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL, nil)
			Expect(err).ToNot(HaveOccurred())

			_, err = testClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
