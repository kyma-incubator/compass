package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	logCtx "github.com/kyma-incubator/compass/components/ias-adapter/internal/logger/context"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Test Suite")
}

func createTestRequest(body any) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	req := &http.Request{
		URL:  &url.URL{},
		Body: io.NopCloser(processBody(body)),
	}
	ctx.Request = req
	requestID := uuid.NewString()
	ctxLogger := logger.Default().With().Str(logCtx.RequestIDCtxKey, requestID).Logger()
	ctx.Set(logCtx.LoggerCtxKey, &ctxLogger)
	ctx.Set(logCtx.RequestIDCtxKey, requestID)
	return w, ctx
}

func processBody(v any) io.Reader {
	if v == nil {
		return nil
	}

	if s, isString := v.(string); isString {
		return strings.NewReader(s)
	}

	data, err := json.Marshal(v)
	Expect(err).ToNot(HaveOccurred())
	return bytes.NewReader(data)
}
