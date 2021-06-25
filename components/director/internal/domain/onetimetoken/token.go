package onetimetoken

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/header"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
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
	url, err := url.Parse(legacyConnectorURL)
	if err != nil {
		return "", errors.Wrapf(err, "while parsing string (%s) as the URL", legacyConnectorURL)
	}

	if url.RawQuery != "" {
		url.RawQuery += "&"
	}
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

func extractTokenFromURL(legacyURLStr string) string {
	legacyURL, err := url.Parse(legacyURLStr)
	if err != nil {
		return ""
	}
	return legacyURL.Query().Get("token")
}
