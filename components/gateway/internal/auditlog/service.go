package auditlog

import (
	"encoding/json"
	"strings"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"
)

type AuditlogMessage struct {
	Request  string
	Response string
	proxy.Claims
}

type AuditlogSource struct {
	logsChannel chan AuditlogMessage
}

func NewAuditlogSink(logsChannel chan AuditlogMessage) *AuditlogSource {
	return &AuditlogSource{
		logsChannel: logsChannel,
	}
}

func (sink *AuditlogSource) Log(request, response string, claims proxy.Claims) error {
	msg := AuditlogMessage{
		Request:  request,
		Response: response,
		Claims:   claims,
	}
	sink.logsChannel <- msg
	return nil
}

//go:generate mockery -name=AuditlogClient -output=automock -outpkg=automock -case=underscore
type AuditlogClient interface {
	LogConfigurationChange(change model.ConfigurationChange) error
	LogSecurityEvent(event model.SecurityEvent) error
}

type AuditlogService struct {
	client AuditlogClient
}

func NewService(client AuditlogClient) *AuditlogService {
	return &AuditlogService{client: client}
}

func (svc *AuditlogService) Log(request, response string, claims proxy.Claims) error {
	graphqlResponse, err := svc.parseResponse(response)
	if err != nil {
		return errors.Wrap(err, "while parsing response")
	}

	if len(graphqlResponse.Errors) == 0 {
		log := svc.createConfigChangeLog(claims, request)
		log.Attributes = append(log.Attributes, model.Attribute{
			Name: "response",
			Old:  "",
			New:  "success",
		})

		err = svc.client.LogConfigurationChange(log)
		return errors.Wrap(err, "while sending to auditlog")
	}

	if idx := svc.isInsufficientScopeError(graphqlResponse.Errors); idx != -1 {
		graphqErr := graphqlResponse.Errors[idx]
		data, err := json.Marshal(&graphqErr)
		if err != nil {
			return errors.Wrap(err, "while marshalling graphql err")
		}

		err = svc.client.LogSecurityEvent(model.SecurityEvent{
			User: "proxy",
			Data: string(data),
		})
		return errors.Wrap(err, "while sending secuity event to auditlog")
	}

	if isReadError(graphqlResponse, request) {
		log := svc.createConfigChangeLog(claims, request)
		log.Attributes = append(log.Attributes, model.Attribute{
			Name: "response",
			Old:  "",
			New:  "success",
		})
		err := svc.client.LogConfigurationChange(log)
		return errors.Wrap(err, "while sending configuration change")
	}

	log := svc.createConfigChangeLog(claims, request)
	log.Attributes = append(log.Attributes, model.Attribute{
		Name: "response",
		Old:  "",
		New:  response,
	})
	err = svc.client.LogConfigurationChange(log)
	return errors.Wrap(err, "while sending configuration change")
}

func (svc *AuditlogService) parseResponse(response string) (GraphqlResponse, error) {
	var graphqResponse GraphqlResponse
	err := json.Unmarshal([]byte(response), &graphqResponse)
	if err != nil {
		return GraphqlResponse{}, err
	}
	return graphqResponse, nil
}

func (svc *AuditlogService) createConfigChangeLog(claims proxy.Claims, request string) model.ConfigurationChange {
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
			{Name: "request", Old: "", New: request}},
	}
}

func (svc *AuditlogService) isInsufficientScopeError(errors []ErrorMessage) int {
	for i, msg := range errors {
		if strings.Contains(msg.Message, "insufficient scopes provided") {
			return i
		}
	}
	return -1
}

type graphqQuery struct {
	Query string `json:"query"`
}

//We assume that if request payload start with `mutation` and
//if any of response errors has path array length equal 1, that means that mutation failed
func isReadError(response GraphqlResponse, request string) bool {
	req := strings.TrimSpace(request)
	mutation := strings.HasPrefix(req, "mutation")
	if !mutation {
		var graphqlQuery graphqQuery
		err := json.Unmarshal([]byte(request), &graphqlQuery)
		if err != nil {
			return true
		}

		graphqlRequestPayload := strings.TrimSpace(graphqlQuery.Query)
		if strings.HasPrefix(graphqlRequestPayload, "mutation") {
			return searchForMutationErr(response)
		}

		return true
	}
	return searchForMutationErr(response)

}

func searchForMutationErr(response GraphqlResponse) bool {
	for _, graphqlErr := range response.Errors {
		if len(graphqlErr.Path) == 1 {
			return false
		}
	}
	return true
}
