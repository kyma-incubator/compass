// Code generated by counterfeiter. DO NOT EDIT.
package httpfakes

import (
	"context"
	"net/url"
	"sync"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/http"
)

type FakeTokenProvider struct {
	GetAuthorizationTokenStub        func(context.Context) (http.Token, error)
	getAuthorizationTokenMutex       sync.RWMutex
	getAuthorizationTokenArgsForCall []struct {
		arg1 context.Context
	}
	getAuthorizationTokenReturns struct {
		result1 http.Token
		result2 error
	}
	getAuthorizationTokenReturnsOnCall map[int]struct {
		result1 http.Token
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
	TargetURLStub        func() *url.URL
	targetURLMutex       sync.RWMutex
	targetURLArgsForCall []struct {
	}
	targetURLReturns struct {
		result1 *url.URL
	}
	targetURLReturnsOnCall map[int]struct {
		result1 *url.URL
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeTokenProvider) GetAuthorizationToken(arg1 context.Context) (http.Token, error) {
	fake.getAuthorizationTokenMutex.Lock()
	ret, specificReturn := fake.getAuthorizationTokenReturnsOnCall[len(fake.getAuthorizationTokenArgsForCall)]
	fake.getAuthorizationTokenArgsForCall = append(fake.getAuthorizationTokenArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	fake.recordInvocation("GetAuthorizationToken", []interface{}{arg1})
	fake.getAuthorizationTokenMutex.Unlock()
	if fake.GetAuthorizationTokenStub != nil {
		return fake.GetAuthorizationTokenStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.getAuthorizationTokenReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeTokenProvider) GetAuthorizationTokenCallCount() int {
	fake.getAuthorizationTokenMutex.RLock()
	defer fake.getAuthorizationTokenMutex.RUnlock()
	return len(fake.getAuthorizationTokenArgsForCall)
}

func (fake *FakeTokenProvider) GetAuthorizationTokenCalls(stub func(context.Context) (http.Token, error)) {
	fake.getAuthorizationTokenMutex.Lock()
	defer fake.getAuthorizationTokenMutex.Unlock()
	fake.GetAuthorizationTokenStub = stub
}

func (fake *FakeTokenProvider) GetAuthorizationTokenArgsForCall(i int) context.Context {
	fake.getAuthorizationTokenMutex.RLock()
	defer fake.getAuthorizationTokenMutex.RUnlock()
	argsForCall := fake.getAuthorizationTokenArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeTokenProvider) GetAuthorizationTokenReturns(result1 http.Token, result2 error) {
	fake.getAuthorizationTokenMutex.Lock()
	defer fake.getAuthorizationTokenMutex.Unlock()
	fake.GetAuthorizationTokenStub = nil
	fake.getAuthorizationTokenReturns = struct {
		result1 http.Token
		result2 error
	}{result1, result2}
}

func (fake *FakeTokenProvider) GetAuthorizationTokenReturnsOnCall(i int, result1 http.Token, result2 error) {
	fake.getAuthorizationTokenMutex.Lock()
	defer fake.getAuthorizationTokenMutex.Unlock()
	fake.GetAuthorizationTokenStub = nil
	if fake.getAuthorizationTokenReturnsOnCall == nil {
		fake.getAuthorizationTokenReturnsOnCall = make(map[int]struct {
			result1 http.Token
			result2 error
		})
	}
	fake.getAuthorizationTokenReturnsOnCall[i] = struct {
		result1 http.Token
		result2 error
	}{result1, result2}
}

func (fake *FakeTokenProvider) Matches(arg1 context.Context) bool {
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

func (fake *FakeTokenProvider) MatchesCallCount() int {
	fake.matchesMutex.RLock()
	defer fake.matchesMutex.RUnlock()
	return len(fake.matchesArgsForCall)
}

func (fake *FakeTokenProvider) MatchesCalls(stub func(context.Context) bool) {
	fake.matchesMutex.Lock()
	defer fake.matchesMutex.Unlock()
	fake.MatchesStub = stub
}

func (fake *FakeTokenProvider) MatchesArgsForCall(i int) context.Context {
	fake.matchesMutex.RLock()
	defer fake.matchesMutex.RUnlock()
	argsForCall := fake.matchesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeTokenProvider) MatchesReturns(result1 bool) {
	fake.matchesMutex.Lock()
	defer fake.matchesMutex.Unlock()
	fake.MatchesStub = nil
	fake.matchesReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeTokenProvider) MatchesReturnsOnCall(i int, result1 bool) {
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

func (fake *FakeTokenProvider) Name() string {
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

func (fake *FakeTokenProvider) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *FakeTokenProvider) NameCalls(stub func() string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *FakeTokenProvider) NameReturns(result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeTokenProvider) NameReturnsOnCall(i int, result1 string) {
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

func (fake *FakeTokenProvider) TargetURL() *url.URL {
	fake.targetURLMutex.Lock()
	ret, specificReturn := fake.targetURLReturnsOnCall[len(fake.targetURLArgsForCall)]
	fake.targetURLArgsForCall = append(fake.targetURLArgsForCall, struct {
	}{})
	fake.recordInvocation("TargetURL", []interface{}{})
	fake.targetURLMutex.Unlock()
	if fake.TargetURLStub != nil {
		return fake.TargetURLStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.targetURLReturns
	return fakeReturns.result1
}

func (fake *FakeTokenProvider) TargetURLCallCount() int {
	fake.targetURLMutex.RLock()
	defer fake.targetURLMutex.RUnlock()
	return len(fake.targetURLArgsForCall)
}

func (fake *FakeTokenProvider) TargetURLCalls(stub func() *url.URL) {
	fake.targetURLMutex.Lock()
	defer fake.targetURLMutex.Unlock()
	fake.TargetURLStub = stub
}

func (fake *FakeTokenProvider) TargetURLReturns(result1 *url.URL) {
	fake.targetURLMutex.Lock()
	defer fake.targetURLMutex.Unlock()
	fake.TargetURLStub = nil
	fake.targetURLReturns = struct {
		result1 *url.URL
	}{result1}
}

func (fake *FakeTokenProvider) TargetURLReturnsOnCall(i int, result1 *url.URL) {
	fake.targetURLMutex.Lock()
	defer fake.targetURLMutex.Unlock()
	fake.TargetURLStub = nil
	if fake.targetURLReturnsOnCall == nil {
		fake.targetURLReturnsOnCall = make(map[int]struct {
			result1 *url.URL
		})
	}
	fake.targetURLReturnsOnCall[i] = struct {
		result1 *url.URL
	}{result1}
}

func (fake *FakeTokenProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getAuthorizationTokenMutex.RLock()
	defer fake.getAuthorizationTokenMutex.RUnlock()
	fake.matchesMutex.RLock()
	defer fake.matchesMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	fake.targetURLMutex.RLock()
	defer fake.targetURLMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeTokenProvider) recordInvocation(key string, args []interface{}) {
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

var _ http.TokenProvider = new(FakeTokenProvider)
