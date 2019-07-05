package director

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func ptrString(s string) *string {
	return &s
}

func ptrFetchMode(fm graphql.FetchMode) *graphql.FetchMode {
	return &fm
}

func ptrCLOB(c graphql.CLOB) *graphql.CLOB {
	return &c
}

func ptrBool(b bool) *bool {
	return &b
}
