// Code generated by counterfeiter. DO NOT EDIT.
package httpfakes

import (
	httpa "net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/http"
)

type FakeHTTPRoundTripper struct {
	RoundTripStub        func(*httpa.Request) (*httpa.Response, error)
	roundTripMutex       sync.RWMutex
	roundTripArgsForCall []struct {
		arg1 *httpa.Request
	}
	roundTripReturns struct {
		result1 *httpa.Response
		result2 error
	}
	roundTripReturnsOnCall map[int]struct {
		result1 *httpa.Response
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeHTTPRoundTripper) RoundTrip(arg1 *httpa.Request) (*httpa.Response, error) {
	fake.roundTripMutex.Lock()
	ret, specificReturn := fake.roundTripReturnsOnCall[len(fake.roundTripArgsForCall)]
	fake.roundTripArgsForCall = append(fake.roundTripArgsForCall, struct {
		arg1 *httpa.Request
	}{arg1})
	stub := fake.RoundTripStub
	fakeReturns := fake.roundTripReturns
	fake.recordInvocation("RoundTrip", []interface{}{arg1})
	fake.roundTripMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeHTTPRoundTripper) RoundTripCallCount() int {
	fake.roundTripMutex.RLock()
	defer fake.roundTripMutex.RUnlock()
	return len(fake.roundTripArgsForCall)
}

func (fake *FakeHTTPRoundTripper) RoundTripCalls(stub func(*httpa.Request) (*httpa.Response, error)) {
	fake.roundTripMutex.Lock()
	defer fake.roundTripMutex.Unlock()
	fake.RoundTripStub = stub
}

func (fake *FakeHTTPRoundTripper) RoundTripArgsForCall(i int) *httpa.Request {
	fake.roundTripMutex.RLock()
	defer fake.roundTripMutex.RUnlock()
	argsForCall := fake.roundTripArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeHTTPRoundTripper) RoundTripReturns(result1 *httpa.Response, result2 error) {
	fake.roundTripMutex.Lock()
	defer fake.roundTripMutex.Unlock()
	fake.RoundTripStub = nil
	fake.roundTripReturns = struct {
		result1 *httpa.Response
		result2 error
	}{result1, result2}
}

func (fake *FakeHTTPRoundTripper) RoundTripReturnsOnCall(i int, result1 *httpa.Response, result2 error) {
	fake.roundTripMutex.Lock()
	defer fake.roundTripMutex.Unlock()
	fake.RoundTripStub = nil
	if fake.roundTripReturnsOnCall == nil {
		fake.roundTripReturnsOnCall = make(map[int]struct {
			result1 *httpa.Response
			result2 error
		})
	}
	fake.roundTripReturnsOnCall[i] = struct {
		result1 *httpa.Response
		result2 error
	}{result1, result2}
}

func (fake *FakeHTTPRoundTripper) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.roundTripMutex.RLock()
	defer fake.roundTripMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeHTTPRoundTripper) recordInvocation(key string, args []interface{}) {
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

var _ http.HTTPRoundTripper = new(FakeHTTPRoundTripper)
