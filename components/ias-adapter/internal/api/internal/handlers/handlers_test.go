package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Test Suite")
}

func createTestRequest(body io.Reader) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	req := &http.Request{
		URL:  &url.URL{},
		Body: io.NopCloser(body),
	}
	ctx.Request = req
	ctxLogger := logger.Default().With().Str("correlationID", uuid.NewString()).Logger()
	ctx.Set("logger", &ctxLogger)
	return w, ctx
}
