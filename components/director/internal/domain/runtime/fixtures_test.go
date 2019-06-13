package runtime_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func fixRuntimePage(runtimes []*model.Runtime) *runtime.RuntimePage {
	return &runtime.RuntimePage{
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

func fixModelRuntime(id, name, description string) *model.Runtime {
	return &model.Runtime{
		ID: id,
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

func fixModelRuntimeWithLabels(id, name string, labels map[string][]string) *model.Runtime {
	return &model.Runtime{
		ID: id,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
		},
		Name:        name,
		Description: nil,
		Labels:      labels,
	}
}

func fixModelRuntimeWithAnnotations(id, name string, annotations map[string]interface{}) *model.Runtime {
	return &model.Runtime{
		ID: id,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
		},
		Name:        name,
		Description: nil,
		Annotations: annotations,
	}
}

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
		Annotations: map[string]interface{}{
			"key": "value",
		},
		Labels: map[string][]string{
			"test": {"val", "val2"},
		},
	}
}

func fixDetailedGQLRuntime(t *testing.T, id, name, description string) *graphql.Runtime {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &graphql.Runtime{
		ID: id,
		Status: &graphql.RuntimeStatus{
			Condition: graphql.RuntimeStatusConditionInitial,
			Timestamp: graphql.Timestamp(time),
		},
		Name:        name,
		Description: &description,
		Tenant:      graphql.Tenant("tenant"),
		Annotations: map[string]interface{}{
			"key": "value",
		},
		Labels: map[string][]string{
			"test": {"val", "val2"},
		},
	}
}

func fixModelRuntimeInput(name, description string) model.RuntimeInput {
	return model.RuntimeInput{
		Name:        name,
		Description: &description,
		Annotations: map[string]interface{}{
			"key": "value",
		},
		Labels: map[string][]string{
			"test": {"val", "val2"},
		},
	}
}

func fixGQLRuntimeInput(name, description string) graphql.RuntimeInput {
	labels := graphql.Labels{
		"test": {"val", "val2"},
	}
	annotations := graphql.Annotations{
		"key": "value",
	}

	return graphql.RuntimeInput{
		Name:        name,
		Description: &description,
		Annotations: &annotations,
		Labels:      &labels,
	}
}
