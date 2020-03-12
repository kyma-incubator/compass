package proxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"

	"github.com/pkg/errors"
)

//go:generate mockery -name=RoundTrip -output=automock -outpkg=automock -case=underscore
type RoundTrip interface {
	RoundTrip(*http.Request) (*http.Response, error)
}

//go:generate mockery -name=AuditlogService -output=automock -outpkg=automock -case=underscore
type AuditlogService interface {
	Log(request, resposne string, claims Claims) error
}

type Transport struct {
	http.RoundTripper
	auditlogSvc AuditlogService
}

func NewTransport(svc AuditlogService, trip RoundTrip) *Transport {
	return &Transport{
		RoundTripper: trip,
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

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, errors.Wrap(err, "while rount trip the request")
	}

	claims, err := t.getClaims(req.Header)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing JWT")
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(responseBody))
	defer httpcommon.CloseBody(resp.Body)

	err = t.auditlogSvc.Log(string(requestBody), string(responseBody), claims)
	if err != nil {
		return nil, errors.Wrap(err, "while sending to auditlog")
	}
	return resp, nil
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
		return Claims{}, errors.New("invalid bearer token")
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
