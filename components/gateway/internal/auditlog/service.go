package auditlog

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"
)

//go:generate mockery --name=MetricCollector --output=automock --outpkg=automock --case=underscore --disable-version-string
type MetricCollector interface {
	SetChannelSize(size int)
}

type Sink struct {
	logsChannel chan proxy.AuditlogMessage
	timeout     time.Duration
	collector   MetricCollector
}

func NewSink(logsChannel chan proxy.AuditlogMessage, timeout time.Duration, collector MetricCollector) *Sink {
	return &Sink{
		logsChannel: logsChannel,
		timeout:     timeout,
		collector:   collector,
	}
}

func (sink *Sink) Log(ctx context.Context, msg proxy.AuditlogMessage) error {
	select {
	case sink.logsChannel <- msg:
		log.C(ctx).Debugf("Successfully registered auditlog message for processing to the queue (size=%d, capacity=%d)",
			len(sink.logsChannel), cap(sink.logsChannel))
		sink.collector.SetChannelSize(len(sink.logsChannel))
	case <-time.After(sink.timeout):
		return errors.New("cannot write to the channel")
	}
	return nil
}

type NoOpService struct {
}

func (sink *NoOpService) Log(context.Context, proxy.AuditlogMessage) error {
	return nil
}

func (sink *NoOpService) PreLog(context.Context, proxy.AuditlogMessage) error {
	return nil
}

const (
	PreAuditlogOperation  = "pre-operation"
	PostAuditlogOperation = "post-operation"
)

//go:generate mockery --name=AuditlogClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type AuditlogClient interface {
	LogConfigurationChange(ctx context.Context, change model.ConfigurationChange) error
	LogSecurityEvent(ctx context.Context, event model.SecurityEvent) error
}

//go:generate mockery --name=AuditlogMessageFactory --output=automock --outpkg=automock --case=underscore --disable-version-string
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

func (svc *Service) PreLog(ctx context.Context, msg proxy.AuditlogMessage) error {
	correlationID := msg.CorrelationIDHeaders[correlation.RequestIDHeaderKey]
	configChangeMsg := svc.createConfigChangeMsg(msg.Claims, msg.Request, correlationID, PreAuditlogOperation)
	err := svc.client.LogConfigurationChange(ctx, configChangeMsg)
	return errors.Wrap(err, "while sending configuration pre-change")
}

func (svc *Service) Log(ctx context.Context, msg proxy.AuditlogMessage) error {
	graphqlResponse, err := svc.parseResponse(msg.Response)
	if err != nil {
		return errors.Wrap(err, "while parsing response")
	}

	correlationID := msg.CorrelationIDHeaders[correlation.RequestIDHeaderKey]

	if len(graphqlResponse.Errors) == 0 {
		configChangeMsg := svc.createConfigChangeMsg(msg.Claims, msg.Request, correlationID, PostAuditlogOperation)
		configChangeMsg.Attributes = append(configChangeMsg.Attributes,
			model.Attribute{
				Name: "response",
				Old:  "",
				New:  "success",
			})

		err = svc.client.LogConfigurationChange(ctx, configChangeMsg)
		return errors.Wrap(err, "while sending to auditlog")
	}

	if svc.hasInsufficientScopeError(graphqlResponse.Errors) {
		securityEventMsg := svc.msgFactory.CreateSecurityEvent()
		eventData := model.SecurityEventData{
			ID:            fillID(msg.Claims, "Security Event"),
			CorrelationID: correlationID,
			Reason:        graphqlResponse.Errors,
		}
		data, err := json.Marshal(&eventData)
		if err != nil {
			return errors.Wrap(err, "while marshalling security event data")
		}

		securityEventMsg.Data = string(data)
		err = svc.client.LogSecurityEvent(ctx, securityEventMsg)
		return errors.Wrap(err, "while sending security event to auditlog")
	}

	isReadErr, err := isReadError(graphqlResponse, msg.Request)
	if err != nil {
		return errors.Wrap(err, "while checking if error is read error")
	}

	configChangeMsg := svc.createConfigChangeMsg(msg.Claims, msg.Request, correlationID, PostAuditlogOperation)
	if isReadErr {
		configChangeMsg.Attributes = append(configChangeMsg.Attributes,
			model.Attribute{
				Name: "response",
				Old:  "",
				New:  "success",
			})
		err := svc.client.LogConfigurationChange(ctx, configChangeMsg)
		return errors.Wrap(err, "while sending configuration change")
	} else {
		configChangeMsg.Attributes = append(configChangeMsg.Attributes,
			model.Attribute{
				Name: "response",
				Old:  "",
				New:  msg.Response,
			})
	}

	err = svc.client.LogConfigurationChange(ctx, configChangeMsg)
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

func (svc *Service) createConfigChangeMsg(claims proxy.Claims, request string, correlationID string, auditlogOperationType string) model.ConfigurationChange {
	msg := svc.msgFactory.CreateConfigurationChange()
	msg.Object = model.Object{ID: fillID(claims, "Config Change")}

	msg.Attributes = []model.Attribute{
		{
			Name: "auditlog_type",
			Old:  "",
			New:  auditlogOperationType,
		},
		{
			Name: "request",
			Old:  "",
			New:  request,
		},
		{
			Name: "correlation_id",
			Old:  "",
			New:  correlationID,
		},
	}

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
