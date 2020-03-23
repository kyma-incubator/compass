package auditlog_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/automock"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestAuditlogService_LogConfigurationChange(t *testing.T) {
	t.Run("Success mutation", func(t *testing.T) {
		//GIVEN
		request := fixRequest()
		response := fixNoErrorResponse(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Unsuccessful mutation", func(t *testing.T) {
		//GIVEN
		request := fixRequest()
		response := fixGraphqlMutationError(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, response)

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Success mutation with read error", func(t *testing.T) {
		//GIVEN
		request := fixRequestWithInvalidQuery()
		response := fixResponseReadError(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Success mutation with multiple read error", func(t *testing.T) {
		//GIVEN
		request := fixRequest()
		response := fixResponseMultipleError(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Unsuccessful mutation wit read error and mutation error", func(t *testing.T) {
		//GIVEN
		request := fixRequest()
		response := fixGraphqlMultiErrorWithMutation(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, response)

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Failed query with error", func(t *testing.T) {
		//GIVEN
		request := fixRequestWithQuery()
		response := fixResponseReadError(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Success mutation with payload as json with read errors", func(t *testing.T) {
		//GIVEN
		request := fixJsonRequest()
		response := fixResponseReadError(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Security event - insufficient scope", func(t *testing.T) {
		//GIVEN
		request := fixRequest()
		graphqlResponse := FixResponseUnsufficientScopes()
		response, err := json.Marshal(&graphqlResponse)
		require.NoError(t, err)
		responseErr, err := json.Marshal(graphqlResponse.Errors)
		require.NoError(t, err)
		claims := fixClaims()
		log := model.SecurityEvent{Data: string(responseErr), User: "proxy"}

		client := &automock.AuditlogClient{}
		client.On("LogSecurityEvent", log).Return(nil)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err = auditlogSvc.Log(request, string(response), claims)

		//THEN
		require.NoError(t, err)
		client.AssertExpectations(t)
	})

	t.Run("Auditlog client return error", func(t *testing.T) {
		//GIVEN
		testError := errors.New("test-error")
		request := fixRequest()
		response := fixNoErrorResponse(t)
		claims := fixClaims()
		log := fixLogSuccess(claims, request, "success")

		client := &automock.AuditlogClient{}
		client.On("LogConfigurationChange", log).Return(testError)
		auditlogSvc := auditlog.NewService(client)

		//WHEN
		err := auditlogSvc.Log(request, response, claims)

		//THEN
		require.Error(t, err)
		assert.Error(t, err, fmt.Sprintf("while sending to auditlog: %s", testError.Error()))
		client.AssertExpectations(t)
	})

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
				Message: "zepsulo sie",
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
				Message: "zepsulo sie",
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

func FixResponseUnsufficientScopes() model.GraphqlResponse {
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
				Message: "zepsulo sie",
				Path:    []interface{}{"registerApplication", "apiDefinition"},
			},
		},
		Data: map[string]string{"value": "value"},
	}
	output, err := json.Marshal(&response)
	require.NoError(t, err)
	return string(output)
}

func fixLogSuccess(claims proxy.Claims, request, response string) model.ConfigurationChange {
	return model.ConfigurationChange{
		User: "proxy",
		Object: model.Object{
			ID: map[string]string{
				"name":           "Config Change",
				"externalTenant": claims.Tenant,
				"apiConsumer":    claims.ConsumerType,
				"consumerID":     claims.ConsumerID,
			},
			Type: "",
		},
		Attributes: []model.Attribute{
			{Name: "request", Old: "", New: request},
			{Name: "response", Old: "", New: response}},
	}
}

func fixResponseMultipleError(t *testing.T) string {
	response := model.GraphqlResponse{
		Errors: []model.ErrorMessage{
			{
				Message: "drugi error",
				Path:    []interface{}{"query", "query"},
			},
			{
				Message: "zepsulo sie",
				Path:    []interface{}{"registerApplication", "apiDefinition"},
			},
		},
		Data: map[string]string{"value": "value"},
	}
	output, err := json.Marshal(&response)
	require.NoError(t, err)
	return string(output)
}
