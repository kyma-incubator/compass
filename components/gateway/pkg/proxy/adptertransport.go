package proxy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
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

	fmt.Println(">>>>>>>>>>>>>>>")

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, req.Body)
	if err != nil {
		panic(err)
	}

	requestBody := buf.Bytes()
	fmt.Println(string(requestBody))
	if err != nil {
		return nil, err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	defer httpcommon.CloseBody(req.Context(), req.Body)

	correlationHeaders := correlation.HeadersForRequest(req)

	preAuditLogger := t.auditlogSvc

	//claims, err := t.getClaims(req.Header)
	//if err != nil {
	//	return nil, errors.Wrap(err, "while parsing JWT")
	//}
	//TODO fix claims
	claims := Claims{
		Tenant:         "",
		ConsumerTenant: "",
		Scopes:         "",
		ConsumerID:     "",
		ConsumerType:   "",
		OnBehalfOf:     "",
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

	fmt.Println(">>>>>>>>>>>>>>>>>> Responce Body: ", string(responseBody))
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