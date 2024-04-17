package apptmpltest

import (
	"errors"
	"fmt"

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

var (
	// NonSelfRegFlowErrorMsg is a test message for non-cert flow
	NonSelfRegFlowErrorMsg = fmt.Sprintf("label %s is forbidden when creating Application Template in a non-cert flow.", TestDistinguishLabel)
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
func SelfRegManagerThatReturnsErrorOnPrep(labels map[string]interface{}) *automock.SelfRegisterManager {
	srm := &automock.SelfRegisterManager{}
	srm.On("IsSelfRegistrationFlow", mock.Anything, labels).Return(true, nil).Once()
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

// SelfRegManagerThatDoesCleanup mock for GetSelfRegDistinguishingLabelKey executed 3 times, PrepareForSelfRegistration once, CleanupSelfRegistration once
func SelfRegManagerThatDoesCleanup(res, labels map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Times(2)
		srm.On("IsSelfRegistrationFlow", mock.Anything, labels).Return(true, nil).Once()
		srm.On("CleanupSelfRegistration", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()
		return srm
	}
}

// SelfRegManagerThatDoesNotCleanupFunc mock for GetSelfRegDistinguishingLabelKey executed once, PrepareForSelfRegistration once
func SelfRegManagerThatDoesNotCleanupFunc(res, labels map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Once()
		srm.On("IsSelfRegistrationFlow", mock.Anything, labels).Return(true, nil).Once()

		return srm
	}
}

// SelfRegManagerCheckIsSelfRegistrationFlowOnce mock for IsSelfRegistrationFlow once
func SelfRegManagerCheckIsSelfRegistrationFlowOnce(labels map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := &automock.SelfRegisterManager{}
		srm.On("IsSelfRegistrationFlow", mock.Anything, labels).Return(false, nil).Once()
		return srm
	}
}

// SelfRegManagerOnlyGetDistinguishedLabelKeyTwice mock for GetSelfRegDistinguishingLabelKey twice
func SelfRegManagerOnlyGetDistinguishedLabelKeyTwice() func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := NoopSelfRegManager()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Twice()
		return srm
	}
}

// SelfRegManagerThatDoesPrepAndInitiatesCleanupButNotFinishIt mock for GetSelfRegDistinguishingLabelKey executed 3 times, PrepareForSelfRegistration once
func SelfRegManagerThatDoesPrepAndInitiatesCleanupButNotFinishIt(res, labels map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := SelfRegManagerThatDoesPrepWithNoErrors(res)()
		srm.On("IsSelfRegistrationFlow", mock.Anything, labels).Return(true, nil).Once()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Times(2)
		return srm
	}
}

// SelfRegManagerThatInitiatesCleanupButNotFinishIt mock for GetSelfRegDistinguishingLabelKey executed 3 times
func SelfRegManagerThatInitiatesCleanupButNotFinishIt(labels map[string]interface{}) func() *automock.SelfRegisterManager {
	return func() *automock.SelfRegisterManager {
		srm := NoopSelfRegManager()
		srm.On("GetSelfRegDistinguishingLabelKey").Return(TestDistinguishLabel).Times(2)
		srm.On("IsSelfRegistrationFlow", mock.Anything, labels).Return(false, nil).Once()
		return srm
	}
}
