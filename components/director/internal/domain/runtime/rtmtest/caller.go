package rtmtest

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/stretchr/testify/mock"
	"testing"
)

func CallerThatDoesNotGetCalled(t *testing.T) *automock.ExternalSvcCaller {
	svcCaller := &automock.ExternalSvcCaller{}
	svcCaller.AssertNotCalled(t, "Call", mock.Anything)
	return svcCaller
}
