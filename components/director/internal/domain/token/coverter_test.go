package token_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/token"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	//GIVEN
	tokenModel := model.OneTimeToken{Token: "token", ConnectorURL: "URL"}
	conv := token.NewConverter()
	//WHEN
	graphqlToken := conv.ToGraphQL(tokenModel)
	//THEN
	assert.Equal(t, "token", graphqlToken.Token)
	assert.Equal(t, "URL", graphqlToken.ConnectorURL)
}
