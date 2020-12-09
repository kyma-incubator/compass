// Code generated by counterfeiter. DO NOT EDIT.
package directorfakes

import (
	"context"
	"sync"

	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/machinebox/graphql"
)

type FakeClient struct {
	DoStub        func(context.Context, *graphql.Request, interface{}) error
	doMutex       sync.RWMutex
	doArgsForCall []struct {
		arg1 context.Context
		arg2 *graphql.Request
		arg3 interface{}
	}
	doReturns struct {
		result1 error
	}
	doReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClient) Do(arg1 context.Context, arg2 *graphql.Request, arg3 interface{}) error {
	fake.doMutex.Lock()
	ret, specificReturn := fake.doReturnsOnCall[len(fake.doArgsForCall)]
	fake.doArgsForCall = append(fake.doArgsForCall, struct {
		arg1 context.Context
		arg2 *graphql.Request
		arg3 interface{}
	}{arg1, arg2, arg3})
	fake.recordInvocation("Do", []interface{}{arg1, arg2, arg3})
	fake.doMutex.Unlock()
	if fake.DoStub != nil {
		return fake.DoStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.doReturns
	return fakeReturns.result1
}

func (fake *FakeClient) DoCallCount() int {
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	return len(fake.doArgsForCall)
}

func (fake *FakeClient) DoCalls(stub func(context.Context, *graphql.Request, interface{}) error) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = stub
}

func (fake *FakeClient) DoArgsForCall(i int) (context.Context, *graphql.Request, interface{}) {
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	argsForCall := fake.doArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeClient) DoReturns(result1 error) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = nil
	fake.doReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeClient) DoReturnsOnCall(i int, result1 error) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = nil
	if fake.doReturnsOnCall == nil {
		fake.doReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.doReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeClient) recordInvocation(key string, args []interface{}) {
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

var _ director.Client = new(FakeClient)
