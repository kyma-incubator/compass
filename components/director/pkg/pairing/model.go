package pairing

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type RequestData struct {
	Application graphql.Application
	Tenant      string
}

type ResponseData struct {
	Token string
}
