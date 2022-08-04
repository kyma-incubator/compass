// Code generated by counterfeiter. DO NOT EDIT.
package controllersfakes

import (
	"context"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
)

type FakeWebhookClient struct {
	DoStub        func(context.Context, *webhookclient.Request) (*webhook.Response, error)
	doMutex       sync.RWMutex
	doArgsForCall []struct {
		arg1 context.Context
		arg2 *webhookclient.Request
	}
	doReturns struct {
		result1 *webhook.Response
		result2 error
	}
	doReturnsOnCall map[int]struct {
		result1 *webhook.Response
		result2 error
	}
	PollStub        func(context.Context, *webhookclient.PollRequest) (*webhook.ResponseStatus, error)
	pollMutex       sync.RWMutex
	pollArgsForCall []struct {
		arg1 context.Context
		arg2 *webhookclient.PollRequest
	}
	pollReturns struct {
		result1 *webhook.ResponseStatus
		result2 error
	}
	pollReturnsOnCall map[int]struct {
		result1 *webhook.ResponseStatus
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeWebhookClient) Do(arg1 context.Context, arg2 *webhookclient.Request) (*webhook.Response, error) {
	fake.doMutex.Lock()
	ret, specificReturn := fake.doReturnsOnCall[len(fake.doArgsForCall)]
	fake.doArgsForCall = append(fake.doArgsForCall, struct {
		arg1 context.Context
		arg2 *webhookclient.Request
	}{arg1, arg2})
	stub := fake.DoStub
	fakeReturns := fake.doReturns
	fake.recordInvocation("Do", []interface{}{arg1, arg2})
	fake.doMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeWebhookClient) DoCallCount() int {
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	return len(fake.doArgsForCall)
}

func (fake *FakeWebhookClient) DoCalls(stub func(context.Context, *webhookclient.Request) (*webhook.Response, error)) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = stub
}

func (fake *FakeWebhookClient) DoArgsForCall(i int) (context.Context, *webhookclient.Request) {
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	argsForCall := fake.doArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeWebhookClient) DoReturns(result1 *webhook.Response, result2 error) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = nil
	fake.doReturns = struct {
		result1 *webhook.Response
		result2 error
	}{result1, result2}
}

func (fake *FakeWebhookClient) DoReturnsOnCall(i int, result1 *webhook.Response, result2 error) {
	fake.doMutex.Lock()
	defer fake.doMutex.Unlock()
	fake.DoStub = nil
	if fake.doReturnsOnCall == nil {
		fake.doReturnsOnCall = make(map[int]struct {
			result1 *webhook.Response
			result2 error
		})
	}
	fake.doReturnsOnCall[i] = struct {
		result1 *webhook.Response
		result2 error
	}{result1, result2}
}

func (fake *FakeWebhookClient) Poll(arg1 context.Context, arg2 *webhookclient.PollRequest) (*webhook.ResponseStatus, error) {
	fake.pollMutex.Lock()
	ret, specificReturn := fake.pollReturnsOnCall[len(fake.pollArgsForCall)]
	fake.pollArgsForCall = append(fake.pollArgsForCall, struct {
		arg1 context.Context
		arg2 *webhookclient.PollRequest
	}{arg1, arg2})
	stub := fake.PollStub
	fakeReturns := fake.pollReturns
	fake.recordInvocation("Poll", []interface{}{arg1, arg2})
	fake.pollMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeWebhookClient) PollCallCount() int {
	fake.pollMutex.RLock()
	defer fake.pollMutex.RUnlock()
	return len(fake.pollArgsForCall)
}

func (fake *FakeWebhookClient) PollCalls(stub func(context.Context, *webhookclient.PollRequest) (*webhook.ResponseStatus, error)) {
	fake.pollMutex.Lock()
	defer fake.pollMutex.Unlock()
	fake.PollStub = stub
}

func (fake *FakeWebhookClient) PollArgsForCall(i int) (context.Context, *webhookclient.PollRequest) {
	fake.pollMutex.RLock()
	defer fake.pollMutex.RUnlock()
	argsForCall := fake.pollArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeWebhookClient) PollReturns(result1 *webhook.ResponseStatus, result2 error) {
	fake.pollMutex.Lock()
	defer fake.pollMutex.Unlock()
	fake.PollStub = nil
	fake.pollReturns = struct {
		result1 *webhook.ResponseStatus
		result2 error
	}{result1, result2}
}

func (fake *FakeWebhookClient) PollReturnsOnCall(i int, result1 *webhook.ResponseStatus, result2 error) {
	fake.pollMutex.Lock()
	defer fake.pollMutex.Unlock()
	fake.PollStub = nil
	if fake.pollReturnsOnCall == nil {
		fake.pollReturnsOnCall = make(map[int]struct {
			result1 *webhook.ResponseStatus
			result2 error
		})
	}
	fake.pollReturnsOnCall[i] = struct {
		result1 *webhook.ResponseStatus
		result2 error
	}{result1, result2}
}

func (fake *FakeWebhookClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.doMutex.RLock()
	defer fake.doMutex.RUnlock()
	fake.pollMutex.RLock()
	defer fake.pollMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeWebhookClient) recordInvocation(key string, args []interface{}) {
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

var _ controllers.WebhookClient = new(FakeWebhookClient)
