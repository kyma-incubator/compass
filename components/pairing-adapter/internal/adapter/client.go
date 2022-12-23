package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
)

//go:generate mockery --name=HTTPDoer --output=automock --outpkg=automock --disable-version-string
type HTTPDoer interface {
	Do(r *http.Request) (*http.Response, error)
}
type ExternalClient struct {
	mapping Mapping
	doer    HTTPDoer
}

func NewClient(doer HTTPDoer, mapping Mapping) *ExternalClient {
	return &ExternalClient{
		doer:    doer,
		mapping: mapping,
	}
}

func (c *ExternalClient) Do(ctx context.Context, app RequestData) (*ExternalToken, error) {
	req, err := c.prepareRequest(app)
	if err != nil {
		return nil, err
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "while performing request")
	}

	defer func() {
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			log.C(ctx).Error("Got error on discarding body content", err)
		}

		if err = resp.Body.Close(); err != nil {
			log.C(ctx).Error("Got error on closing response body", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("wrong status code, got: [%d], cannot read body: [%v]", resp.StatusCode, err)
		}

		log.C(ctx).Infof("Wrong status code but got response: %s\n", string(b))
		return nil, fmt.Errorf("wrong status code, got: [%d], body: [%s]", resp.StatusCode, string(b))
	}

	tkn, err := c.getTokenFromResponse(ctx, resp.Body)

	if err != nil {
		return nil, err
	}

	return &ExternalToken{Token: tkn}, nil
}

func (c *ExternalClient) prepareRequest(reqData RequestData) (*http.Request, error) {
	finalURL, err := c.getURL(reqData)
	if err != nil {
		return nil, err
	}
	body, err := c.getBody(reqData)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, finalURL, body)
	if err != nil {
		return nil, err
	}

	headers, err := c.getHeaders(reqData)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		for _, singleHeader := range v {
			request.Header.Add(k, singleHeader)
		}
	}

	return request, nil
}

func (c *ExternalClient) getURL(reqData RequestData) (string, error) {
	urlTemplate, err := template.New("url").Parse(c.mapping.TemplateExternalURL)
	if err != nil {
		return "", err
	}
	finalURL := new(bytes.Buffer)

	err = urlTemplate.Execute(finalURL, reqData)
	if err != nil {
		return "", err
	}
	return finalURL.String(), nil
}

func (c *ExternalClient) getBody(reqData RequestData) (io.Reader, error) {
	bodyTemplate, err := template.New("body").Parse(c.mapping.TemplateJSONBody)
	body := new(bytes.Buffer)
	if err != nil {
		return nil, err
	}

	// temporary code that needs to be removed after T13B 2022 has reached Live
	for idx, value := range reqData.ScenarioGroups {
		if value == "ALL_COMMUNICATION_SCENARIOS" {
			fmt.Println("dummy print")
			reqData.ScenarioGroups[idx] = "UNRESTRICTED"
		}
	}

	if err := bodyTemplate.Execute(body, reqData); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *ExternalClient) getHeaders(reqData RequestData) (map[string][]string, error) {
	headerTemplate, err := template.New("header").Parse(c.mapping.TemplateHeaders)
	if err != nil {
		return nil, err
	}

	finalHeaders := new(bytes.Buffer)

	if err := headerTemplate.Execute(finalHeaders, reqData); err != nil {
		return nil, err
	}

	if finalHeaders.Len() == 0 {
		return nil, nil
	}
	h := map[string][]string{}
	err = json.Unmarshal(finalHeaders.Bytes(), &h)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling headers from JSON to map")
	}
	return h, nil
}

func (c *ExternalClient) getTokenFromResponse(ctx context.Context, in io.Reader) (string, error) {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return "", errors.Wrap(err, "while reading response body")
	}
	respBody := map[string]interface{}{}
	if err = json.Unmarshal(b, &respBody); err != nil {
		return "", errors.Wrap(err, "while unmarshalling response body")
	}

	log.C(ctx).Infof("Got response: %s\n", string(b))
	respTpl, err := template.New("response").Option("missingkey=error").Parse(c.mapping.TemplateTokenFromResponse)
	if err != nil {
		return "", err
	}
	out := new(bytes.Buffer)
	if err := respTpl.Execute(out, respBody); err != nil {
		return "", err
	}

	return out.String(), nil

}
