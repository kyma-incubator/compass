package rtmtest

import (
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/stretchr/testify/mock"
)

const (
	TestDistinguishLabel = "test-distinguish-label"
	SelfRegErrorMsg      = "error during self-reg prep"
)

func NoopSelfRegManager() *automock.SelfRegisterManager {
	return &automock.SelfRegisterManager{}
}

func SelfRegManagerThatDoesPrepWithNoErrors() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("PrepareRuntimeForSelfRegistration", mock.Anything, mock.Anything).Return(nil).Once()
	return srm
}

func SelfRegManagerThatReturnsErrorOnPrep() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("PrepareRuntimeForSelfRegistration", mock.Anything, mock.Anything).Return(errors.New(SelfRegErrorMsg)).Once()
	return srm
}

func SelfRegManagerThatDoesCleanupWithNoErrors() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
	return srm
}

func SelfRegManagerThatReturnsErrorOnCleanup() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string")).Return(errors.New(SelfRegErrorMsg)).Once()
	return srm
}

func SelfRegManagerReturnsDistinguishingLabel() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	return srm
}

func SelfRegManagerThatReturnsNoErrors() *automock.SelfRegisterManager {
	srm := SelfRegManagerThatDoesPrepWithNoErrors()
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
	return srm
}
