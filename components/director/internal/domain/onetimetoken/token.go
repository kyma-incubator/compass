package onetimetoken

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/header"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func raw(obj *graphql.TokenWithURL) (*string, error) {
	rawJSON, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling object to JSON")
	}

	rawJSONStr := string(rawJSON)

	return &rawJSONStr, nil
}

func rawEncoded(obj *graphql.TokenWithURL) (*string, error) {
	rawJSON, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling object to JSON")
	}

	rawBaseEncoded := base64.StdEncoding.EncodeToString(rawJSON)

	return &rawBaseEncoded, nil
}

func legacyConnectorUrlWithToken(legacyConnectorURL, token string) (string, error) {
	if strings.Contains(token, legacyConnectorURL) {
		return token, nil
	}
	url, err := url.Parse(legacyConnectorURL)
	if err != nil {
		return "", errors.Wrapf(err, "while parsing string (%s) as the URL", legacyConnectorURL)
	}

	if url.RawQuery != "" {
		url.RawQuery += "&"
	}
	token = extractToken(token)
	url.RawQuery += fmt.Sprintf("token=%s", token)
	return url.String(), nil
}

func tokenSuggestionEnabled(ctx context.Context, suggestTokenHeaderKey string) bool {
	reqHeaders, ok := ctx.Value(header.ContextKey).(http.Header)
	if !ok || reqHeaders.Get(suggestTokenHeaderKey) != "true" {
		return false
	}

	log.C(ctx).Infof("Token suggestion is required by client")
	return true
}

func extractToken(currentToken string) string {
	// already encoded
	if rawEncoded, err := base64.StdEncoding.DecodeString(currentToken); err == nil {
		if token := gjson.GetBytes(rawEncoded, "token").String(); token != "" {
			return token
		}
	}

	// is in form of a legacy connector url
	legacyURL, err := url.Parse(currentToken)
	if err != nil {
		return currentToken
	}
	if token := legacyURL.Query().Get("token"); token != "" {
		return token
	}

	return currentToken
}
