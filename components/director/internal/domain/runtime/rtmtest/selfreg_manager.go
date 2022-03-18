package rtmtest

import (
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/stretchr/testify/mock"
)

const (
	// TestDistinguishLabel is a test distinguishing label key
	TestDistinguishLabel = "test-distinguish-label"
	// SelfRegErrorMsg is a test error message
	SelfRegErrorMsg = "error during self-reg prep"
)

// NoopSelfRegManager missing godoc
func NoopSelfRegManager() *automock.SelfRegisterManager {
	return &automock.SelfRegisterManager{}
}

// SelfRegManagerThatDoesPrepWithNoErrors missing godoc
func SelfRegManagerThatDoesPrepWithNoErrors(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := &automock.SelfRegisterManager{}
		srm.On("PrepareRuntimeForSelfRegistration", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(res, nil).Once()
		return srm
	}
}

// SelfRegManagerThatReturnsErrorOnPrep missing godoc
func SelfRegManagerThatReturnsErrorOnPrep() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	labels := make(map[string]interface{})
	srm.On("PrepareRuntimeForSelfRegistration", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(labels, errors.New(SelfRegErrorMsg)).Once()
	return srm
}

// SelfRegManagerThatDoesCleanupWithNoErrors missing godoc
func SelfRegManagerThatDoesCleanupWithNoErrors() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()
	return srm
}

// SelfRegManagerThatDoesNotCleanup missing godoc
func SelfRegManagerThatDoesNotCleanup() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	return srm
}

// SelfRegManagerThatReturnsErrorOnCleanup missing godoc
func SelfRegManagerThatReturnsErrorOnCleanup() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(errors.New(SelfRegErrorMsg)).Once()
	return srm
}

// SelfRegManagerReturnsDistinguishingLabel missing godoc
func SelfRegManagerReturnsDistinguishingLabel() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	return srm
}

// SelfRegManagerThatReturnsNoErrors missing godoc
func SelfRegManagerThatReturnsNoErrors(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
		srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()
		return srm
	}
}

// SelfRegManagerThatDoesNotCleanupFunc missing godoc
func SelfRegManagerThatDoesNotCleanupFunc(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
		return srm
	}
}

// SelfRegManagerThatFailsOnCleanup missing godoc
func SelfRegManagerThatFailsOnCleanup(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
		srm.On("CleanupSelfRegisteredRuntime", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(TestError).Once()
		return srm
	}
}
