package apptmpltest

import (
	"errors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"
	"github.com/stretchr/testify/mock"
)

const (
	// TestDistinguishLabel is a test distinguishing label key
	TestDistinguishLabel = "test-distinguish-label"
	// SelfRegErrorMsg is a test error message
	SelfRegErrorMsg = "error during self-reg prep"
)

// NoopSelfRegManager is a noop mock
func NoopSelfRegManager() *automock.SelfRegisterManager {
	return &automock.SelfRegisterManager{}
}

// SelfRegManagerThatDoesPrepWithNoErrors mock for PrepareForSelfRegistration executed once
func SelfRegManagerThatDoesPrepWithNoErrors(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := &automock.SelfRegisterManager{}
		srm.On("PrepareForSelfRegistration", mock.Anything, resource.ApplicationTemplate, mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(res, nil).Once()
		return srm
	}
}

// SelfRegManagerThatReturnsErrorOnPrep mock for GetSelfRegDistinguishingLabelKey executed once with error
func SelfRegManagerThatReturnsErrorOnPrep() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	labels := make(map[string]interface{})
	srm.On("PrepareForSelfRegistration", mock.Anything, resource.ApplicationTemplate, mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(labels, errors.New(SelfRegErrorMsg)).Once()
	return srm
}

// SelfRegManagerThatReturnsErrorOnPrepAndGetSelfRegDistinguishingLabelKey mock for PrepareForSelfRegistration executed once with error, GetSelfRegDistinguishingLabelKey once
func SelfRegManagerThatReturnsErrorOnPrepAndGetSelfRegDistinguishingLabelKey() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	labels := make(map[string]interface{})
	srm.On("PrepareForSelfRegistration", mock.Anything, resource.ApplicationTemplate, mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(labels, errors.New(SelfRegErrorMsg)).Once()
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	return srm
}

// SelfRegManagerThatDoesCleanupWithNoErrors mock for GetSelfRegDistinguishingLabelKey executed once, CleanupSelfRegistration once
func SelfRegManagerThatDoesCleanupWithNoErrors() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	srm.On("CleanupSelfRegistration", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()
	return srm
}

// SelfRegManagerThatReturnsErrorOnCleanup mock for GetSelfRegDistinguishingLabelKey executed once, CleanupSelfRegistration once with error
func SelfRegManagerThatReturnsErrorOnCleanup() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	srm.On("CleanupSelfRegistration", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(errors.New(SelfRegErrorMsg)).Once()
	return srm
}

// SelfRegManagerReturnsDistinguishingLabel mock for GetSelfRegDistinguishingLabelKey executed once
func SelfRegManagerReturnsDistinguishingLabel() *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
	return srm
}

// SelfRegManagerThatDoesCleanup mock for GetSelfRegDistinguishingLabelKey executed 2 times, PrepareForSelfRegistration once, CleanupSelfRegistration once
func SelfRegManagerThatDoesCleanup(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Times(2)
		srm.On("CleanupSelfRegistration", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()
		return srm
	}
}

// SelfRegManagerThatDoesNotCleanupFunc mock for GetSelfRegDistinguishingLabelKey executed once, PrepareForSelfRegistration once
func SelfRegManagerThatDoesNotCleanupFunc(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
		return srm
	}
}

// SelfRegManagerThatInitiatesCleanupButNotFinishIt mock for GetSelfRegDistinguishingLabelKey executed 2 times, PrepareForSelfRegistration once
func SelfRegManagerThatInitiatesCleanupButNotFinishIt(res map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Times(2)
		return srm
	}
}
