package ucl

import (
	"context"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUCLService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UCL Service Test Suite")
}

var _ = Describe("Reporting status", func() {
	var ctx context.Context = context.Background()

	It("Errors When json marshal fails", func() {
		service := NewService(&http.Client{})
		err := service.ReportStatus(ctx, "url", StatusReport{Configuration: make(chan int)})
		Expect(err.Error()).To(HavePrefix("failed to json marshal status report "))
	})

	It("Errors When request creation fails", func() {
		service := NewService(&http.Client{})
		var nilCtx context.Context = nil
		err := service.ReportStatus(nilCtx, "invalid-url", StatusReport{})
		Expect(err.Error()).To(HavePrefix("failed to create request:"))
	})

	It("Errors When request execution fails", func() {
		service := NewService(&http.Client{})
		err := service.ReportStatus(ctx, "invalid-url", StatusReport{})
		Expect(err.Error()).To(HavePrefix("failed to execute request:"))
	})

	It("Errors When status is not 200 and body is invalid", func() {
		service := NewService(&http.Client{Transport: &testTransport{statusCode: http.StatusInternalServerError}})
		err := service.ReportStatus(ctx, "https://valid.url", StatusReport{})
		Expect(err.Error()).To(HavePrefix("unexpected response status 500, body:"))
	})

	It("Succeeds When status is 200", func() {
		service := NewService(&http.Client{Transport: &testTransport{statusCode: http.StatusOK}})
		err := service.ReportStatus(ctx, "https://valid.url", StatusReport{})
		Expect(err).ToNot(HaveOccurred())
	})
})

type testTransport struct {
	err        error
	statusCode int
}

func (tt *testTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if tt.err != nil {
		return nil, tt.err
	}

	return &http.Response{
		StatusCode: tt.statusCode,
		Status:     http.StatusText(tt.statusCode),
	}, nil
}
