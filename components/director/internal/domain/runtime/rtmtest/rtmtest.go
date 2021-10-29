package rtmtest

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/stretchr/testify/mock"
)

const TestDistinguishLabel = "test-distinguish-label"

func NoopSelfRegManager() *automock.SelfRegisterManager {
	return &automock.SelfRegisterManager{}
}

func SelfRegManagerThatReturnsNoErrors() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("PrepareRuntimeForSelfRegistration", mock.Anything, mock.Anything).Return(nil).Once()
	srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	return srm
}
