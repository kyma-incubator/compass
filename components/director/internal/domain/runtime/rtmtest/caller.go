package rtmtest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/stretchr/testify/mock"
)

const (
	ResponseLabelKey   = "test-label-key"
	ResponseLabelValue = "test-label-val"
)

var TestError = errors.New("test-error")

func CallerThatDoesNotGetCalled(t *testing.T) *automock.ExternalSvcCaller {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.AssertNotCalled(t, "Call", mock.Anything)
	return svcCaller
}

func CallerThatDoesNotSucceed(*testing.T) *automock.ExternalSvcCaller {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.On("Call", mock.Anything).Return(nil, TestError).Once()
	return svcCaller
}

func CallerThatReturnsBadStatus(*testing.T) *automock.ExternalSvcCaller {
	svcCaller := &automock.ExternalSvcCaller{}
	response := httptest.ResponseRecorder{
		Code: http.StatusBadRequest,
		Body: bytes.NewBufferString(""),
	}
	svcCaller.On("Call", mock.Anything).Return(response.Result(), nil).Once()
	return svcCaller
}

func CallerThatGetsCalledOnce(statusCode int) func(*testing.T) *automock.ExternalSvcCaller {
	return func(t *testing.T) *automock.ExternalSvcCaller {
		svcCaller := &automock.ExternalSvcCaller{}
		response := httptest.ResponseRecorder{
			Code: statusCode,
			Body: bytes.NewBufferString(fmt.Sprintf(`{"%s":"%s"}`, ResponseLabelKey, ResponseLabelValue)),
		}
		svcCaller.On("Call", mock.Anything).Return(response.Result(), nil).Once()
		return svcCaller
	}
}
