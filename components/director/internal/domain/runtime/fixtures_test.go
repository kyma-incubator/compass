package runtime_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func fixRuntimePage(runtimes []*model.Runtime) *model.RuntimePage {
	return &model.RuntimePage{
		Data: runtimes,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(runtimes),
	}
}

func fixGQLRuntimePage(runtimes []*graphql.Runtime) *graphql.RuntimePage {
	return &graphql.RuntimePage{
		Data: runtimes,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(runtimes),
	}
}

func fixModelRuntime(id, tenant, name, description string) *model.Runtime {
	return &model.Runtime{
		ID:     id,
		Tenant: tenant,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
	}
}

func fixGQLRuntime(id, name, description string) *graphql.Runtime {
	return &graphql.Runtime{
		ID: id,
		Status: &graphql.RuntimeStatus{
			Condition: graphql.RuntimeStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
	}
}

//func fixModelRuntimeWithLabels(id, tenant, name string, labels map[string]interface{}) *model.Runtime {
//	return &model.Runtime{
//		ID:     id,
//		Tenant: tenant,
//		Status: &model.RuntimeStatus{
//			Condition: model.RuntimeStatusConditionInitial,
//		},
//		Name:        name,
//		Description: nil,
//	}
//}

func fixDetailedModelRuntime(t *testing.T, id, name, description string) *model.Runtime {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &model.Runtime{
		ID: id,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
			Timestamp: time,
		},
		Name:        name,
		Description: &description,
		Tenant:      "tenant",
		AgentAuth: &model.Auth{
			AdditionalHeaders: map[string][]string{
				"test": {"bar"},
			},
			Credential: model.CredentialData{
				Basic: &model.BasicCredentialData{
					Username: "foo",
					Password: "bar",
				},
			},
		},
	}
}

func fixDetailedGQLRuntime(t *testing.T, id, name, description string) *graphql.Runtime {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	headers := graphql.HttpHeaders{
		"test": {"bar"},
	}

	return &graphql.Runtime{
		ID: id,
		Status: &graphql.RuntimeStatus{
			Condition: graphql.RuntimeStatusConditionInitial,
			Timestamp: graphql.Timestamp(time),
		},
		Name:        name,
		Description: &description,
		AgentAuth: &graphql.Auth{
			AdditionalHeaders: &headers,
			Credential: graphql.BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
	}
}

func fixModelRuntimeInput(name, description string) model.RuntimeInput {
	return model.RuntimeInput{
		Name:        name,
		Description: &description,
		Labels: map[string]interface{}{
			"test": []string{"val", "val2"},
		},
	}
}

func fixGQLRuntimeInput(name, description string) graphql.RuntimeInput {
	labels := graphql.Labels{
		"test": []string{"val", "val2"},
	}

	return graphql.RuntimeInput{
		Name:        name,
		Description: &description,
		Labels:      &labels,
	}
}
