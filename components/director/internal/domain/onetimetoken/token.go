package onetimetoken

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
