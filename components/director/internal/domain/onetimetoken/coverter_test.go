package onetimetoken_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	validConnectorURL                string = "http://connector.domain.local"
	validLegacyConnectorURL          string = "http://legacy-connector.domain.local"
	validLegacyConnectorURLWithQuery string = "http://legacy-connector.domain.local?abc=1"
	invalidURL                       string = ":123:invalid-url"
	token                            string = "token123"
)

func TestConverter_ToGraphQLForRuntime(t *testing.T) {
	//GIVEN
	tokenModel := model.OneTimeToken{Token: token, ConnectorURL: validConnectorURL}
	conv := onetimetoken.NewConverter(validLegacyConnectorURL)
	//WHEN
	graphqlToken := conv.ToGraphQLForRuntime(tokenModel)
	//THEN
	assert.Equal(t, token, graphqlToken.Token)
	assert.Equal(t, validConnectorURL, graphqlToken.ConnectorURL)
}

func TestConverter_ToGraphQLForApplication(t *testing.T) {
	//GIVEN
	tokenModel := model.OneTimeToken{Token: token, ConnectorURL: validConnectorURL}
	conv := onetimetoken.NewConverter(validLegacyConnectorURL)
	//WHEN
	graphqlToken, err := conv.ToGraphQLForApplication(tokenModel)
	//THEN
	assert.NoError(t, err)
	assert.Equal(t, token, graphqlToken.Token)
	assert.Equal(t, validConnectorURL, graphqlToken.ConnectorURL)
	assert.Equal(t, fmt.Sprintf("%s?token=%s", validLegacyConnectorURL, token), graphqlToken.LegacyConnectorURL)
}

func TestConverter_ToGraphQLForApplication_WithLegacyURLWithQueryParam(t *testing.T) {
	//GIVEN
	tokenModel := model.OneTimeToken{Token: token, ConnectorURL: validConnectorURL}
	conv := onetimetoken.NewConverter(validLegacyConnectorURLWithQuery)
	// WHEN
	graphqlToken, err := conv.ToGraphQLForApplication(tokenModel)
	//THEN
	assert.NoError(t, err)
	assert.Equal(t, token, graphqlToken.Token)
	assert.Equal(t, validConnectorURL, graphqlToken.ConnectorURL)
	assert.Equal(t, fmt.Sprintf("%s&token=%s", validLegacyConnectorURLWithQuery, token), graphqlToken.LegacyConnectorURL)
}

func TestConverter_ToGraphQLForApplication_ErrorWithInvalidLegacyURL(t *testing.T) {
	//GIVEN
	tokenModel := model.OneTimeToken{Token: token, ConnectorURL: validConnectorURL}
	conv := onetimetoken.NewConverter(invalidURL)
	// WHEN
	_, err := conv.ToGraphQLForApplication(tokenModel)
	//THEN
	assert.EqualError(t, err, "while parsing string (:123:invalid-url) as the URL: parse \":123:invalid-url\": missing protocol scheme")
}
