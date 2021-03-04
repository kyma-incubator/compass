// Code generated by counterfeiter. DO NOT EDIT.
package httpfakes

import (
	"context"
	"sync"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/http"
)

type FakeAuthorizationProvider struct {
	GetAuthorizationStub        func(context.Context) (string, error)
	getAuthorizationMutex       sync.RWMutex
	getAuthorizationArgsForCall []struct {
		arg1 context.Context
	}
	getAuthorizationReturns struct {
		result1 string
		result2 error
	}
	getAuthorizationReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	MatchesStub        func(context.Context) bool
	matchesMutex       sync.RWMutex
	matchesArgsForCall []struct {
		arg1 context.Context
	}
	matchesReturns struct {
		result1 bool
	}
	matchesReturnsOnCall map[int]struct {
		result1 bool
	}
	NameStub        func() string
	nameMutex       sync.RWMutex
	nameArgsForCall []struct {
	}
	nameReturns struct {
		result1 string
	}
	nameReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeAuthorizationProvider) GetAuthorization(arg1 context.Context) (string, error) {
	fake.getAuthorizationMutex.Lock()
	ret, specificReturn := fake.getAuthorizationReturnsOnCall[len(fake.getAuthorizationArgsForCall)]
	fake.getAuthorizationArgsForCall = append(fake.getAuthorizationArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	fake.recordInvocation("GetAuthorization", []interface{}{arg1})
	fake.getAuthorizationMutex.Unlock()
	if fake.GetAuthorizationStub != nil {
		return fake.GetAuthorizationStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getAuthorizationReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeAuthorizationProvider) GetAuthorizationCallCount() int {
	fake.getAuthorizationMutex.RLock()
	defer fake.getAuthorizationMutex.RUnlock()
	return len(fake.getAuthorizationArgsForCall)
}

func (fake *FakeAuthorizationProvider) GetAuthorizationCalls(stub func(context.Context) (string, error)) {
	fake.getAuthorizationMutex.Lock()
	defer fake.getAuthorizationMutex.Unlock()
	fake.GetAuthorizationStub = stub
}

func (fake *FakeAuthorizationProvider) GetAuthorizationArgsForCall(i int) context.Context {
	fake.getAuthorizationMutex.RLock()
	defer fake.getAuthorizationMutex.RUnlock()
	argsForCall := fake.getAuthorizationArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeAuthorizationProvider) GetAuthorizationReturns(result1 string, result2 error) {
	fake.getAuthorizationMutex.Lock()
	defer fake.getAuthorizationMutex.Unlock()
	fake.GetAuthorizationStub = nil
	fake.getAuthorizationReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeAuthorizationProvider) GetAuthorizationReturnsOnCall(i int, result1 string, result2 error) {
	fake.getAuthorizationMutex.Lock()
	defer fake.getAuthorizationMutex.Unlock()
	fake.GetAuthorizationStub = nil
	if fake.getAuthorizationReturnsOnCall == nil {
		fake.getAuthorizationReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getAuthorizationReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeAuthorizationProvider) Matches(arg1 context.Context) bool {
	fake.matchesMutex.Lock()
	ret, specificReturn := fake.matchesReturnsOnCall[len(fake.matchesArgsForCall)]
	fake.matchesArgsForCall = append(fake.matchesArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	fake.recordInvocation("Matches", []interface{}{arg1})
	fake.matchesMutex.Unlock()
	if fake.MatchesStub != nil {
		return fake.MatchesStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.matchesReturns
	return fakeReturns.result1
}

func (fake *FakeAuthorizationProvider) MatchesCallCount() int {
	fake.matchesMutex.RLock()
	defer fake.matchesMutex.RUnlock()
	return len(fake.matchesArgsForCall)
}

func (fake *FakeAuthorizationProvider) MatchesCalls(stub func(context.Context) bool) {
	fake.matchesMutex.Lock()
	defer fake.matchesMutex.Unlock()
	fake.MatchesStub = stub
}

func (fake *FakeAuthorizationProvider) MatchesArgsForCall(i int) context.Context {
	fake.matchesMutex.RLock()
	defer fake.matchesMutex.RUnlock()
	argsForCall := fake.matchesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeAuthorizationProvider) MatchesReturns(result1 bool) {
	fake.matchesMutex.Lock()
	defer fake.matchesMutex.Unlock()
	fake.MatchesStub = nil
	fake.matchesReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeAuthorizationProvider) MatchesReturnsOnCall(i int, result1 bool) {
	fake.matchesMutex.Lock()
	defer fake.matchesMutex.Unlock()
	fake.MatchesStub = nil
	if fake.matchesReturnsOnCall == nil {
		fake.matchesReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.matchesReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeAuthorizationProvider) Name() string {
	fake.nameMutex.Lock()
	ret, specificReturn := fake.nameReturnsOnCall[len(fake.nameArgsForCall)]
	fake.nameArgsForCall = append(fake.nameArgsForCall, struct {
	}{})
	fake.recordInvocation("Name", []interface{}{})
	fake.nameMutex.Unlock()
	if fake.NameStub != nil {
		return fake.NameStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.nameReturns
	return fakeReturns.result1
}

func (fake *FakeAuthorizationProvider) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *FakeAuthorizationProvider) NameCalls(stub func() string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *FakeAuthorizationProvider) NameReturns(result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeAuthorizationProvider) NameReturnsOnCall(i int, result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeAuthorizationProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getAuthorizationMutex.RLock()
	defer fake.getAuthorizationMutex.RUnlock()
	fake.matchesMutex.RLock()
	defer fake.matchesMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeAuthorizationProvider) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ http.AuthorizationProvider = new(FakeAuthorizationProvider)
