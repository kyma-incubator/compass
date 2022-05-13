package proxy

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"
	"github.com/pkg/errors"
)

// AdapterConfig stores the configuration for the adapter transport
type AdapterConfig struct {
	MsgBodySizeLimit int
}

type AdapterTransport struct {
	config AdapterConfig
	http.RoundTripper
	auditlogSink AuditlogService
	auditlogSvc  PreAuditlogService
}

func NewAdapterTransport(sink AuditlogService, svc PreAuditlogService, trip RoundTrip, config AdapterConfig) *AdapterTransport {
	return &AdapterTransport{
		RoundTripper: trip,
		auditlogSink: sink,
		auditlogSvc:  svc,
		config:       config,
	}
}

type auditLogContext struct {
	preAuditLog        PreAuditlogService
	postAuditLog       AuditlogService
	cfg                AdapterConfig
	correlationHeaders correlation.Headers
	claims             Claims
	logId              uuid.UUID
}

func newAuditLogContext(t *AdapterTransport, req *http.Request) (*auditLogContext, error) {
	correlationHeaders := correlation.HeadersForRequest(req)

	claims, err := getClaims(req.Context(), req.Header)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing JWT")
	}

	transactionId, err := uuid.NewRandom()
	if nil != err {
		return nil, errors.Wrap(err, "Failed to generate transaction UUID")
	}

	return &auditLogContext{
		preAuditLog:        t.auditlogSvc,
		postAuditLog:       t.auditlogSink,
		cfg:                t.config,
		correlationHeaders: correlationHeaders,
		claims:             claims,
		logId:              transactionId,
	}, nil
}

func (t *AdapterTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.Body == nil || req.Method == http.MethodGet {
		return t.RoundTripper.RoundTrip(req)
	}
	defer httpcommon.CloseBody(req.Context(), req.Body)

	buffer := &bytes.Buffer{}
	_, err = io.Copy(buffer, req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(buffer)

	auditLogCtx, err := newAuditLogContext(t, req)
	if nil != err {
		return nil, err
	}

	ctx := context.WithValue(req.Context(), correlation.RequestIDHeaderKey, auditLogCtx.correlationHeaders)

	if err := doPreAuditLog(ctx, auditLogCtx, buffer.Bytes()); nil != err {
		return nil, errors.Wrap(err, "Failed to do pre-change auditlog")
	}

	log.C(ctx).Info("Proxying request...")
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, errors.Wrap(err, "on request round trip")
	}
	defer httpcommon.CloseBody(req.Context(), resp.Body)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(responseBody))

	if err := doPostAuditLog(ctx, auditLogCtx, responseBody); nil != err {
		log.C(ctx).WithError(err).Error("failed to send post-change audit-log")
	}

	return resp, nil
}

func doPreAuditLog(ctx context.Context, auditLogCtx *auditLogContext, requestBody []byte) error {
	log.C(ctx).Infof("Maximum request body length per audit log message: %d", auditLogCtx.cfg.MsgBodySizeLimit)
	log.C(ctx).Infof("Length of received request body: %d", len(requestBody))

	transactionId := auditLogCtx.logId.String()
	shardLength := calculateShardLength(ctx, auditLogCtx, requestBody, transactionId)

	log.C(ctx).Infof("Writing pre-change auditlog for %s...", transactionId)
	for len(requestBody) > 0 {
		length := min(shardLength, len(requestBody))
		shard := requestBody[:length]
		requestBody = requestBody[length:]

		err := auditLogCtx.preAuditLog.PreLog(ctx, AuditlogMessage{
			CorrelationIDHeaders: auditLogCtx.correlationHeaders,
			Request:              transactionId + ": " + string(shard),
			Response:             "",
			Claims:               auditLogCtx.claims,
		})

		if nil != err {
			return errors.Wrapf(err, "while sending pre-change auditlog message to auditlog service for %s", transactionId)
		}
	}

	return nil
}

func doPostAuditLog(ctx context.Context, auditLogCtx *auditLogContext, responseBody []byte) error {
	log.C(ctx).Infof("Length of received response body: %d", len(responseBody))

	transactionId := auditLogCtx.logId.String()
	shardLength := calculateShardLength(ctx, auditLogCtx, responseBody, transactionId)

	log.C(ctx).Infof("Writing post-change auditlog for %s...", transactionId)
	for len(responseBody) > 0 {
		length := min(shardLength, len(responseBody))
		shard := responseBody[:length]
		responseBody = responseBody[length:]

		err := auditLogCtx.postAuditLog.Log(ctx, AuditlogMessage{
			CorrelationIDHeaders: auditLogCtx.correlationHeaders,
			Request:              "",
			Response:             transactionId + ": " + string(shard),
			Claims:               auditLogCtx.claims,
		})

		if nil != err {
			return errors.Wrapf(err, "while sending post-change auditlog message to auditlog service for %s", transactionId)
		}
	}

	return nil
}

func calculateShardLength(ctx context.Context, auditLogCtx *auditLogContext, bodyBytes []byte, transactionId string) int {
	shards := len(bodyBytes) / auditLogCtx.cfg.MsgBodySizeLimit
	if len(bodyBytes)%auditLogCtx.cfg.MsgBodySizeLimit != 0 {
		shards += 1
	}

	// Compute the maximum size of each auditlog message
	shardLength := min(auditLogCtx.cfg.MsgBodySizeLimit, len(bodyBytes))
	if shards > 1 {
		shardLength = len(bodyBytes) / shards
	}

	log.C(ctx).Infof("The data for %s will be split to %d shards of approximately %d bytes", transactionId, shards, shardLength)

	return shardLength
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

func min(a int, b int) int {
	if a <= b {
		return a
	}

	return b
}
