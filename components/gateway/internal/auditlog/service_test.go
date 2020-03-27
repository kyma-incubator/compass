package auditlog_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/automock"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditlogService_LogConfigurationChange(t *testing.T) {
	t.Run("Success mutation", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		request := fixRequest()
		response := fixNoErrorResponse(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Unsuccessful mutation", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		request := fixRequest()
		response := fixGraphqlMutationError(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, response)

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Success mutation with read error", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		request := fixRequestWithInvalidQuery()
		response := fixResponseReadError(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Success mutation with multiple read error", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		request := fixRequest()
		response := fixResponseMultipleError(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Unsuccessful mutation wit read error and mutation error", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		request := fixRequest()
		response := fixGraphqlMultiErrorWithMutation(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, response)

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Failed query with error", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		request := fixRequestWithQuery()
		response := fixResponseReadError(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Success mutation with payload as json with read errors", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		request := fixJsonRequest()
		response := fixResponseReadError(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Security event - insufficient scope", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateSecurityEvent").Return(fixFabricatedSecurityEventMsg())

		request := fixRequest()
		graphqlResponse := fixResponseUnsufficientScopes()
		response, err := json.Marshal(&graphqlResponse)
		require.NoError(t, err)

		claims := fixClaims()
		msg := fixSecurityEventMsg(t, graphqlResponse.Errors, fixClaims())

		client := &automock.AuditlogClient{}
		client.On("LogSecurityEvent", msg).Return(nil)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err = auditlogSvc.Log(request, string(response), claims)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, client, factory)
	})

	t.Run("Auditlog client return error", func(t *testing.T) {
		//GIVEN
		factory := &automock.AuditlogMessageFactory{}
		factory.On("CreateConfigurationChange").Return(fixFabricatedConfigChangeMsg())

		testError := errors.New("test-error")
		request := fixRequest()
		response := fixNoErrorResponse(t)
		claims := fixClaims()
		log := fixSuccessConfigChangeMsg(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(testError)
		auditlogSvc := auditlog.NewService(client, factory)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.Error(t, err)
		assert.Error(t, err, fmt.Sprintf("while sending to auditlog: %s", testError.Error()))
		mock.AssertExpectationsForObjects(t, client, factory)
	})

}
func TestSink_ChannelStuck(t *testing.T) {
	//GIVEN
	chanMsg := make(chan auditlog.Message)
	defer close(chanMsg)
	sink := auditlog.NewSink(chanMsg, time.Millisecond*100)

	//WHEN
	err := sink.Log("test-request", "test-response", proxy.Claims{})

	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, "Cannot write to the channel")
}

func fixClaims() proxy.Claims {
	return proxy.Claims{
		Tenant:       "e36c520b-caa2-4677-b289-8a171184192b",
		Scopes:       "scopes",
		ConsumerID:   "134039be-840a-47f1-a962-d13410edf311",
		ConsumerType: "Application",
	}
}

func fixRequest() string {
	return `
		mutation {
		   registerApplication(in : {name:"test"}) {
		  id
		  name
		  }
		  registerRuntime(in: {name:"app2"}) {
		  id
		  name
		  }
		}`
}

func fixJsonRequest() string {
	return `{"query":"mutation a{\n   registerApplication(in : {name:\"test123\"}) {\n  id\n  name\n  labels\n  apiDefinition(id:\"\") {\n    id\n  }\n  }\n  registerRuntime(in: {name:\"app2\"}) {\n      id\n  name\n  labels\n  }\n}\n","operationName":"a"}`
}

func fixRequestWithQuery() string {
	return `query wiever {
			  result: viewer {
				id
				type
			  }
			}`
}

func fixRequestWithInvalidQuery() string {
	return `mutation {
			   registerApplication(in : {name:"test1"}) {
			  id
			  name
			  labels
			  apiDefinition(id:"") {
				id
			  }
			  }
			  registerRuntime(in: {name:"app2"}) {
				  id
			  name
			  labels
			  }
			}`
}

func fixGraphqlMutationError(t *testing.T) string {
	response := model.GraphqlResponse{
		Errors: []model.ErrorMessage{
			{
				Message: "first error",
				Path:    []interface{}{"registerRuntime"},
			},
		},
		Data: map[string]string{
			"value": "value",
		},
	}
	output, err := json.Marshal(&response)
	require.NoError(t, err)
	return string(output)
}

func fixGraphqlMultiErrorWithMutation(t *testing.T) string {
	response := model.GraphqlResponse{
		Errors: []model.ErrorMessage{
			{
				Message: "first error",
				Path:    []interface{}{"registerRuntime"},
			},
			{
				Message: "read error",
				Path:    []interface{}{"queyr", "failed"},
			},
		},
		Data: map[string]string{
			"value": "value",
		},
	}
	output, err := json.Marshal(&response)
	require.NoError(t, err)
	return string(output)
}

func fixResponseUnsufficientScopes() model.GraphqlResponse {
	return model.GraphqlResponse{
		Errors: []model.ErrorMessage{
			{
				Message: "insufficient scopes provided, required: [application:write], actual: []",
				Path:    []interface{}{"path", "path"},
			},
		},
		Data: map[string]string{
			"value": "value",
		},
	}
}

func fixNoErrorResponse(t *testing.T) string {
	response := model.GraphqlResponse{
		Errors: nil,
		Data: map[string]string{
			"value": "value",
		},
	}
	output, err := json.Marshal(&response)
	require.NoError(t, err)
	return string(output)
}

func fixResponseReadError(t *testing.T) string {
	response := model.GraphqlResponse{
		Errors: []model.ErrorMessage{
			{
				Message: "first error",
				Path:    []interface{}{"registerApplication", "apiDefinition"},
			},
		},
		Data: map[string]string{"value": "value"},
	}
	output, err := json.Marshal(&response)
	require.NoError(t, err)
	return string(output)
}

func fixResponseMultipleError(t *testing.T) string {
	response := model.GraphqlResponse{
		Errors: []model.ErrorMessage{
			{
				Message: "first error",
				Path:    []interface{}{"query", "query"},
			},
			{
				Message: "second error",
				Path:    []interface{}{"registerApplication", "apiDefinition"},
			},
		},
		Data: map[string]string{"value": "value"},
	}
	output, err := json.Marshal(&response)
	require.NoError(t, err)
	return string(output)
}
