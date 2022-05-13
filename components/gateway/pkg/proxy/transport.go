package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"
	"github.com/pkg/errors"
)

const (
	consumerTenant = "consumerTenant"
)

var emptyQuery error = errors.New("empty graphql query")

//go:generate mockery --name=RoundTrip --output=automock --outpkg=automock --case=underscore --disable-version-string
type RoundTrip interface {
	RoundTrip(*http.Request) (*http.Response, error)
}

//go:generate mockery --name=AuditlogService --output=automock --outpkg=automock --case=underscore --disable-version-string
type AuditlogService interface {
	Log(ctx context.Context, msg AuditlogMessage) error
}

//go:generate mockery --name=PreAuditlogService --output=automock --outpkg=automock --case=underscore --disable-version-string
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
	auditlogSvc  PreAuditlogService
}

func NewTransport(sink AuditlogService, svc PreAuditlogService, trip RoundTrip) *Transport {
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
	defer httpcommon.CloseBody(req.Context(), req.Body)

	correlationHeaders := correlation.HeadersForRequest(req)

	isMutation, err := checkQueryType(requestBody, "mutation")
	if err != nil && err != emptyQuery {
		return nil, errors.Wrap(err, "could not check query type")
	}

	if !isMutation && err != emptyQuery {
		log.C(req.Context()).Debugln("Will not send auditlog message for queries")
		return t.RoundTripper.RoundTrip(req)
	}

	preAuditLogger := t.auditlogSvc

	claims, err := getClaims(req.Context(), req.Header)
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
	defer httpcommon.CloseBody(req.Context(), resp.Body)

	err = t.auditlogSink.Log(req.Context(), AuditlogMessage{
		CorrelationIDHeaders: correlationHeaders,
		Request:              string(requestBody),
		Response:             string(responseBody),
		Claims:               claims,
	})
	if err != nil {
		log.C(ctx).WithError(err).Errorf("failed to send a post-change auditlog message to auditlog service: %v", err)
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
	Tenant         string `json:"tenant"`
	ConsumerTenant string `json:"consumerTenant"`
	Scopes         string `json:"scopes"`
	ConsumerID     string `json:"consumerID"`
	ConsumerType   string `json:"consumerType"`
	OnBehalfOf     string `json:"onBehalfOf"`
}

func (c Claims) Valid() error {
	return nil
}

func getClaims(ctx context.Context, headers http.Header) (Claims, error) {
	tokenClaims := struct {
		TenantString string `json:"tenant"`
		Scopes       string `json:"scopes"`
		ConsumerID   string `json:"consumerID"`
		ConsumerType string `json:"consumerType"`
		OnBehalfOf   string `json:"onBehalfOf"`
		jwt.StandardClaims
	}{}

	token := headers.Get("Authorization")
	if token == "" {
		return Claims{}, errors.New("no bearer token")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	parser := jwt.Parser{SkipClaimsValidation: true}

	_, _, err := parser.ParseUnverified(token, &tokenClaims)

	if err != nil {
		return Claims{}, errors.Wrap(err, "while parsing bearer token")
	}

	tenants := make(map[string]string, 0)
	err = json.Unmarshal([]byte(tokenClaims.TenantString), &tenants)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("failed to extract tenants from token, proceeding with empty tenants: %v", err)
	}

	claims := Claims{
		Tenant:         tenants[consumerTenant],
		ConsumerTenant: tenants[consumerTenant],
		Scopes:         tokenClaims.Scopes,
		ConsumerID:     tokenClaims.ConsumerID,
		ConsumerType:   tokenClaims.ConsumerType,
		OnBehalfOf:     tokenClaims.OnBehalfOf,
	}

	if provider, exists := tenants["providerTenant"]; exists {
		claims.Tenant = provider
	}

	return claims, nil
}
