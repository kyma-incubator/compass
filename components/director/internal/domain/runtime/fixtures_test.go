package runtime_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
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
			//Timestamp: //TODO: Wait for scalar marshalling
		},
		Name:        name,
		Description: &description,
		// Tenant:      in.Tenant, //TODO: Wait for scalar marshalling
		//Annotations:in.Annotations, //TODO: Wait for scalar marshalling
		//Labels:in.Labels, //TODO: Wait for scalar marshalling
	}
}

func fixGQLRuntime(id, name, description string) *graphql.Runtime {
	return &graphql.Runtime{
		ID: id,
		Status: &graphql.RuntimeStatus{
			Condition: graphql.RuntimeStatusConditionInitial,
			//Timestamp: //TODO: Wait for scalar marshalling
		},
		Name:        name,
		Description: &description,
		// Tenant:      in.Tenant, //TODO: Wait for scalar marshalling
		//Annotations:in.Annotations, //TODO: Wait for scalar marshalling
		//Labels:in.Labels, //TODO: Wait for scalar marshalling
	}
}

func fixModelRuntimeInput(name, description string) model.RuntimeInput {
	return model.RuntimeInput{
		Name:        name,
		Description: &description,
		//Annotations:in.Annotations, //TODO: Wait for scalar marshalling
		//Labels:in.Labels, //TODO: Wait for scalar marshalling
	}
}

func fixGQLRuntimeInput(name, description string) graphql.RuntimeInput {
	return graphql.RuntimeInput{
		Name:        name,
		Description: &description,
		//Annotations:in.Annotations, //TODO: Wait for scalar marshalling
		//Labels:in.Labels, //TODO: Wait for scalar marshalling
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
		Labels: labels,
	}
}

func fixModelRuntimeWithAnnotations(id, name string, annotations map[string]string) *model.Runtime {
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