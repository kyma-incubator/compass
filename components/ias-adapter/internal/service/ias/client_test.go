package ias

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logCtx "github.com/kyma-incubator/compass/components/ias-adapter/internal/logger/context"
)

var _ = Describe("IAS Client", func() {
	When("Outbound request is sent", func() {
		It("Request ID header should be present", func() {
			testClient := &http.Client{
				Transport: &headerTransport{clientTransport: http.DefaultTransport},
			}
			requestID := "123"
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Header.Get(logCtx.RequestIDHeader)).To(Equal(requestID))
			}))
			defer testServer.Close()

			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Set(logCtx.RequestIDCtxKey, requestID)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL, nil)
			Expect(err).ToNot(HaveOccurred())

			_, err = testClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
