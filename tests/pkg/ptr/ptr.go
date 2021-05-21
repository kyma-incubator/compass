package ptr

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func String(s string) *string {
	return &s
}

func FetchMode(fm graphql.FetchMode) *graphql.FetchMode {
	return &fm
}

func CLOB(c graphql.CLOB) *graphql.CLOB {
	return &c
}

func Bool(b bool) *bool {
	return &b
}
