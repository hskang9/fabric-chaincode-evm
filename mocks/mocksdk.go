// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
)

type MockSDK struct {
	GetChannelClientStub        func() (fabproxy.ChannelClient, error)
	getChannelClientMutex       sync.RWMutex
	getChannelClientArgsForCall []struct{}
	getChannelClientReturns     struct {
		result1 fabproxy.ChannelClient
		result2 error
	}
	getChannelClientReturnsOnCall map[int]struct {
		result1 fabproxy.ChannelClient
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *MockSDK) GetChannelClient() (fabproxy.ChannelClient, error) {
	fake.getChannelClientMutex.Lock()
	ret, specificReturn := fake.getChannelClientReturnsOnCall[len(fake.getChannelClientArgsForCall)]
	fake.getChannelClientArgsForCall = append(fake.getChannelClientArgsForCall, struct{}{})
	fake.recordInvocation("GetChannelClient", []interface{}{})
	fake.getChannelClientMutex.Unlock()
	if fake.GetChannelClientStub != nil {
		return fake.GetChannelClientStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.getChannelClientReturns.result1, fake.getChannelClientReturns.result2
}

func (fake *MockSDK) GetChannelClientCallCount() int {
	fake.getChannelClientMutex.RLock()
	defer fake.getChannelClientMutex.RUnlock()
	return len(fake.getChannelClientArgsForCall)
}

func (fake *MockSDK) GetChannelClientReturns(result1 fabproxy.ChannelClient, result2 error) {
	fake.GetChannelClientStub = nil
	fake.getChannelClientReturns = struct {
		result1 fabproxy.ChannelClient
		result2 error
	}{result1, result2}
}

func (fake *MockSDK) GetChannelClientReturnsOnCall(i int, result1 fabproxy.ChannelClient, result2 error) {
	fake.GetChannelClientStub = nil
	if fake.getChannelClientReturnsOnCall == nil {
		fake.getChannelClientReturnsOnCall = make(map[int]struct {
			result1 fabproxy.ChannelClient
			result2 error
		})
	}
	fake.getChannelClientReturnsOnCall[i] = struct {
		result1 fabproxy.ChannelClient
		result2 error
	}{result1, result2}
}

func (fake *MockSDK) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getChannelClientMutex.RLock()
	defer fake.getChannelClientMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *MockSDK) recordInvocation(key string, args []interface{}) {
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

var _ fabproxy.SDK = new(MockSDK)