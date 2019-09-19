package token

import (
	"context"
	"fmt"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var requestForRuntime = `
		mutation { generateRuntimeToken (runtimeID:"%s")
		  {
			token
		  }
		}`

//go:generate mockery -name=GraphQLClient -output=automock -outpkg=automock -case=underscore
type GraphQLClient interface{
	Run(ctx context.Context, req *gcli.Request, resp interface{}) error
}

type TokenService struct {
	cli GraphQLClient
}

func NewTokenService(gcli GraphQLClient) *TokenService {
	return &TokenService{cli: gcli}
}

func (t TokenService) GetOneTimeTokenForRuntime(ctx context.Context, id string) (string, error) {
	request := newRuntimeTokenRequest(requestForRuntime, id)
	output := ExternalTokenModel{}
	err := t.cli.Run(ctx, request, &output)
	if err != nil {
		return "", errors.Wrap(err, "while calling connector for runtime one time token")
	}
	return output.GenerateRuntimeToken.Token, err
}

func newRuntimeTokenRequest(request, id string) *gcli.Request {
	return gcli.NewRequest(fmt.Sprintf(request, id))
}
