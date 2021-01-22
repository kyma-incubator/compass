package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"
	"github.com/pkg/errors"
)

var emptyQuery error = errors.New("empty graphql query")

//go:generate mockery --name=RoundTrip --output=automock --outpkg=automock --case=underscore
type RoundTrip interface {
	RoundTrip(*http.Request) (*http.Response, error)
}

//go:generate mockery --name=AuditlogService --output=automock --outpkg=automock --case=underscore
type AuditlogService interface {
	Log(ctx context.Context, msg AuditlogMessage) error
}

//go:generate mockery --name=PreAuditlogService --output=automock --outpkg=automock --case=underscore
type PreAuditlogService interface {
	AuditlogService
	PreLog(ctx context.Context, msg AuditlogMessage) error
}

type AuditlogMessage struct {
	CorrelationIDHeaders correlation.Headers
	Request              string
	Response             string
	Claims
}

type Transport struct {
	http.RoundTripper
	auditlogSink AuditlogService
	auditlogSvc  AuditlogService
}

func NewTransport(sink AuditlogService, svc AuditlogService, trip RoundTrip) *Transport {
	return &Transport{
		RoundTripper: trip,
		auditlogSink: sink,
		auditlogSvc:  svc,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.Body == nil || req.Method == http.MethodGet {
		return t.RoundTripper.RoundTrip(req)
	}

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	defer httpcommon.CloseBody(req.Body)

	correlationHeaders := correlation.HeadersForRequest(req)

	isMutation, err := checkQueryType(requestBody, "mutation")
	if err != nil && err != emptyQuery {
		return nil, errors.Wrap(err, "could not check query type")
	}

	if !isMutation && err != emptyQuery {
		log.Println("Will not send auditlog message for queries")
		return t.RoundTripper.RoundTrip(req)
	}

	preAuditLogger, ok := t.auditlogSvc.(PreAuditlogService)
	if !ok {
		return nil, errors.New("Failed to type cast PreAuditlogService")
	}

	claims, err := t.getClaims(req.Header)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing JWT")
	}

	ctx := context.WithValue(req.Context(), correlation.RequestIDHeaderKey, correlationHeaders)
	err = preAuditLogger.PreLog(ctx, AuditlogMessage{
		CorrelationIDHeaders: correlationHeaders,
		Request:              string(requestBody),
		Response:             "",
		Claims:               claims,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while sending pre-change auditlog message to auditlog service")
	}

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, errors.Wrap(err, "on request round trip")
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(responseBody))
	defer httpcommon.CloseBody(resp.Body)

	err = t.auditlogSink.Log(req.Context(), AuditlogMessage{
		CorrelationIDHeaders: correlationHeaders,
		Request:              string(requestBody),
		Response:             string(responseBody),
		Claims:               claims,
	})
	if err != nil {
		log.Printf("failed to send a post-change auditlog message to auditlog service: %v", err)
	}

	return resp, nil
}

func checkQueryType(requestBody []byte, typee string) (bool, error) {
	var query map[string]interface{}
	if err := json.Unmarshal(requestBody, &query); err != nil {
		return false, errors.Wrap(err, "could not unmarshal query")
	}
	queryObj, ok := query["query"]
	if !ok {
		return false, emptyQuery
	}
	queryString, ok := queryObj.(string)
	if !ok {
		return false, errors.New("query is not a string")
	}
	queryString = strings.TrimSpace(queryString)
	if strings.HasPrefix(queryString, typee) {
		return true, nil
	}
	return false, nil
}

type Claims struct {
	Tenant       string `json:"tenant"`
	Scopes       string `json:"scopes"`
	ConsumerID   string `json:"consumerID"`
	ConsumerType string `json:"consumerType"`
}

func (c Claims) Valid() error {
	return nil
}

func (t *Transport) getClaims(headers http.Header) (Claims, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return Claims{}, errors.New("no bearer token")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	parser := jwt.Parser{SkipClaimsValidation: true}
	claims := Claims{}
	_, _, err := parser.ParseUnverified(token, &claims)
	if err != nil {
		return claims, errors.Wrap(err, "while parsing beaerer token")
	}
	return claims, nil
}
