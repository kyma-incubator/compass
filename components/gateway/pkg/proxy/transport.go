package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/kyma-incubator/compass/components/gateway/pkg/httpcommon"

	"github.com/pkg/errors"
)

var emptyQuery error = errors.New("empty graphql query")

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

	isMutation, err := checkQueryType(requestBody, "mutation")
	if err != emptyQuery {
		if err != nil {
			return nil, err
		}
		if !isMutation {
			log.Println("Will not auditlog queries")
			return resp, nil
		}
	}

	err = t.auditlogSvc.Log(string(requestBody), string(responseBody), claims)
	if err != nil {
		return nil, errors.Wrap(err, "while sending to auditlog")
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

	if strings.HasPrefix(string(queryString), typee) {
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
