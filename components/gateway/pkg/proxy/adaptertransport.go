package proxy

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"
	"github.com/pkg/errors"
)

type AdapterTransport struct {
	http.RoundTripper
	auditlogSink AuditlogService
	auditlogSvc  PreAuditlogService
}

func NewAdapterTransport(sink AuditlogService, svc PreAuditlogService, trip RoundTrip) *AdapterTransport {
	return &AdapterTransport{
		RoundTripper: trip,
		auditlogSink: sink,
		auditlogSvc:  svc,
	}
}

func (t *AdapterTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.Body == nil || req.Method == http.MethodGet {
		return t.RoundTripper.RoundTrip(req)
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, req.Body)
	if err != nil {
		return nil, err
	}

	requestBody := buf.Bytes()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	defer httpcommon.CloseBody(req.Context(), req.Body)

	correlationHeaders := correlation.HeadersForRequest(req)

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

type AdapterTokenClaims struct {
	ExtraAttributes map[string]interface{} `json:"ext_attr"`
	Scopes          []string               `json:"scope"`
	ConsumerID      string                 `json:"client_id"`
	jwt.StandardClaims
}

func (t *AdapterTransport) getClaims(headers http.Header) (Claims, error) {
	tokenClaims := AdapterTokenClaims{}
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

	claims := Claims{
		Tenant:     tokenClaims.ExtraAttributes["subaccountid"].(string),
		Scopes:     strings.Join(tokenClaims.Scopes, ", "),
		ConsumerID: tokenClaims.ConsumerID,
	}

	return claims, nil
}
