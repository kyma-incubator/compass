package auditlog

import (
	"encoding/json"
	"strings"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"
)

type Message struct {
	Request  string
	Response string
	proxy.Claims
}

type Sink struct {
	logsChannel chan Message
}

func NewSink(logsChannel chan Message) *Sink {
	return &Sink{
		logsChannel: logsChannel,
	}
}

func (sink *Sink) Log(request, response string, claims proxy.Claims) error {
	msg := Message{
		Request:  request,
		Response: response,
		Claims:   claims,
	}
	sink.logsChannel <- msg
	return nil
}

type NoOpService struct {
}

func (sink *NoOpService) Log(request, response string, claims proxy.Claims) error {
	return nil
}

//go:generate mockery -name=AuditlogClient -output=automock -outpkg=automock -case=underscore
type AuditlogClient interface {
	LogConfigurationChange(change model.ConfigurationChange) error
	LogSecurityEvent(event model.SecurityEvent) error
}

//go:generate mockery -name=AuditlogMessageFactory -output=automock -outpkg=automock -case=underscore
type AuditlogMessageFactory interface {
	CreateConfigurationChange() model.ConfigurationChange
	CreateSecurityEvent() model.SecurityEvent
}

type Service struct {
	client     AuditlogClient
	msgFactory AuditlogMessageFactory
}

func NewService(client AuditlogClient, msgFactory AuditlogMessageFactory) *Service {
	return &Service{client: client, msgFactory: msgFactory}
}

func (svc *Service) Log(request, response string, claims proxy.Claims) error {
	graphqlResponse, err := svc.parseResponse(response)
	if err != nil {
		return errors.Wrap(err, "while parsing response")
	}

	if len(graphqlResponse.Errors) == 0 {
		log := svc.createConfigChangeMsg(claims, request)
		log.Attributes = append(log.Attributes, model.Attribute{
			Name: "response",
			Old:  "",
			New:  "success",
		})

		err = svc.client.LogConfigurationChange(log)
		return errors.Wrap(err, "while sending to auditlog")
	}

	if svc.hasInsufficientScopeError(graphqlResponse.Errors) {
		response, err := json.Marshal(&graphqlResponse.Errors)
		if err != nil {
			return errors.Wrap(err, "while marshalling graphql err")
		}

		msg := svc.msgFactory.CreateSecurityEvent()
		eventData := model.SecurityEventData{ID: fillID(claims, "Security Event"), Reason: string(response)}
		data, err := json.Marshal(&eventData)
		if err != nil {
			return errors.Wrap(err, "while marshalling security event data")
		}

		msg.Data = string(data)
		err = svc.client.LogSecurityEvent(msg)
		return errors.Wrap(err, "while sending security event to auditlog")
	}

	isReadErr, err := isReadError(graphqlResponse, request)
	if err != nil {
		return errors.Wrap(err, "while checking if error is read error")
	}

	msg := svc.createConfigChangeMsg(claims, request)
	if isReadErr {
		msg.Attributes = append(msg.Attributes, model.Attribute{
			Name: "response",
			Old:  "",
			New:  "success",
		})
		err := svc.client.LogConfigurationChange(msg)
		return errors.Wrap(err, "while sending configuration change")
	} else {
		msg.Attributes = append(msg.Attributes, model.Attribute{
			Name: "response",
			Old:  "",
			New:  response,
		})
	}

	err = svc.client.LogConfigurationChange(msg)
	return errors.Wrap(err, "while sending configuration change")
}

func (svc *Service) parseResponse(response string) (model.GraphqlResponse, error) {
	var graphqlResponse model.GraphqlResponse
	err := json.Unmarshal([]byte(response), &graphqlResponse)
	if err != nil {
		return model.GraphqlResponse{}, err
	}
	return graphqlResponse, nil
}

func (svc *Service) createConfigChangeMsg(claims proxy.Claims, request string) model.ConfigurationChange {
	msg := svc.msgFactory.CreateConfigurationChange()
	msg.Object = model.Object{ID: fillID(claims, "Config Change")}
	msg.Attributes = []model.Attribute{
		{Name: "request", Old: "", New: request}}

	return msg
}

func (svc *Service) hasInsufficientScopeError(errors []model.ErrorMessage) bool {
	for _, msg := range errors {
		if strings.Contains(msg.Message, "insufficient scopes provided") {
			return true
		}
	}
	return false
}

type graphqQuery struct {
	Query string `json:"query"`
}

//We assume that if request payload start with `mutation` and
//if any of response errors has path array length equal 1, that means that mutation failed
func isReadError(response model.GraphqlResponse, request string) (bool, error) {
	req := strings.TrimSpace(request)
	isMutation := strings.HasPrefix(req, "mutation")
	if isMutation {
		return searchForMutationErr(response), nil
	}

	isQuery := strings.HasPrefix(req, "query")
	if isQuery {
		return true, nil
	}

	var graphqlQuery graphqQuery
	err := json.Unmarshal([]byte(request), &graphqlQuery)
	if err != nil {
		return false, errors.Wrap(err, "while marshalling graphql named query")
	}

	graphqlRequestPayload := strings.TrimSpace(graphqlQuery.Query)
	if strings.HasPrefix(graphqlRequestPayload, "mutation") {
		return searchForMutationErr(response), nil
	}

	return true, nil
}

func searchForMutationErr(response model.GraphqlResponse) bool {
	for _, graphqlErr := range response.Errors {
		if len(graphqlErr.Path) == 1 {
			return false
		}
	}
	return true
}

func fillID(claims proxy.Claims, name string) map[string]string {
	return map[string]string{
		"name":           name,
		"externalTenant": claims.Tenant,
		"apiConsumer":    claims.ConsumerType,
		"consumerID":     claims.ConsumerID,
	}
}
