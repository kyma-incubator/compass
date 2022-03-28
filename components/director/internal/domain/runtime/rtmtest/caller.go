package rtmtest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/config"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/stretchr/testify/mock"
)

const (
	// ResponseLabelKey is a testing label key
	ResponseLabelKey = "test-response-label-key"
	// ResponseLabelValue is a testing label value
	ResponseLabelValue = "test-label-val"
)

// TestError is a testing error
var TestError = errors.New("test-error")

// CallerThatDoesNotGetCalled missing godoc
func CallerThatDoesNotGetCalled(t *testing.T, _ config.SelfRegConfig, _ string) *automock.ExternalSvcCallerProvider {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.AssertNotCalled(t, "Call", mock.Anything)

	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.AssertNotCalled(t, "GetCaller", mock.Anything, mock.Anything)
	return svcCallerProvider
}

// CallerThatDoesNotSucceed missing godoc
func CallerThatDoesNotSucceed(_ *testing.T, cfg config.SelfRegConfig, region string) *automock.ExternalSvcCallerProvider {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.On("Call", mock.Anything).Return(nil, TestError).Once()

	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
	return svcCallerProvider
}

// CallerThatReturnsBadStatus missing godoc
func CallerThatReturnsBadStatus(_ *testing.T, cfg config.SelfRegConfig, region string) *automock.ExternalSvcCallerProvider {
	svcCaller := &automock.ExternalSvcCaller{}
	response := httptest.ResponseRecorder{
		Code: http.StatusBadRequest,
		Body: bytes.NewBufferString(""),
	}
	svcCaller.On("Call", mock.Anything).Return(response.Result(), nil).Once()

	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
	return svcCallerProvider
}

// CallerThatGetsCalledOnce missing godoc
func CallerThatGetsCalledOnce(statusCode int) func(*testing.T, config.SelfRegConfig, string) *automock.ExternalSvcCallerProvider {
	return func(t *testing.T, cfg config.SelfRegConfig, region string) *automock.ExternalSvcCallerProvider {
		svcCaller := &automock.ExternalSvcCaller{}
		response := httptest.ResponseRecorder{
			Code: statusCode,
			Body: bytes.NewBufferString(fmt.Sprintf(`{"%s":"%s"}`, ResponseLabelKey, ResponseLabelValue)),
		}
		svcCaller.On("Call", mock.Anything).Return(response.Result(), nil).Once()

		svcCallerProvider := &automock.ExternalSvcCallerProvider{}
		svcCallerProvider.On("GetCaller", cfg, region).Return(svcCaller, nil).Once()
		return svcCallerProvider
	}
}

// CallerProviderThatFails missing godoc
func CallerProviderThatFails(_ *testing.T, cfg config.SelfRegConfig, region string) *automock.ExternalSvcCallerProvider {
	svcCallerProvider := &automock.ExternalSvcCallerProvider{}
	svcCallerProvider.On("GetCaller", cfg, region).Return(nil, TestError).Once()
	return svcCallerProvider
}
