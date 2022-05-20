package runtime_test

import (
	"net/url"
	"testing"
	"time"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/require"
)

const (
	tenantID  = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	runtimeID = "runtimeID"
)

var fixColumns = []string{"id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}

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

func fixModelRuntime(t *testing.T, id, tenant, name, description string) *model.Runtime {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &model.Runtime{
		ID: id,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
		},
		Name:              name,
		Description:       &description,
		CreationTimestamp: time,
	}
}

func fixGQLRuntime(t *testing.T, id, name, description string) *graphql.Runtime {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &graphql.Runtime{
		ID: id,
		Status: &graphql.RuntimeStatus{
			Condition: graphql.RuntimeStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
		Metadata: &graphql.RuntimeMetadata{
			CreationTimestamp: graphql.Timestamp(time),
		},
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
		Name:              name,
		Description:       &description,
		CreationTimestamp: time,
	}
}

func fixDetailedEntityRuntime(t *testing.T, id, name, description string) *runtime.Runtime {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &runtime.Runtime{
		ID:                id,
		StatusCondition:   string(model.RuntimeStatusConditionInitial),
		StatusTimestamp:   time,
		Name:              name,
		Description:       repo.NewValidNullableString(description),
		CreationTimestamp: time,
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
		Metadata: &graphql.RuntimeMetadata{
			CreationTimestamp: graphql.Timestamp(time),
		},
	}
}

func fixModelRuntimeRegisterInput(name, description string, webhooks []*model.WebhookInput) model.RuntimeRegisterInput {
	return model.RuntimeRegisterInput{
		Name:        name,
		Description: &description,
		Labels: map[string]interface{}{
			"test": []string{"val", "val2"},
		},
		Webhooks: webhooks,
	}
}

func fixGQLRuntimeRegisterInput(name, description string, webhooks []*graphql.WebhookInput) graphql.RuntimeRegisterInput {
	labels := graphql.Labels{
		"test": []string{"val", "val2"},
	}

	return graphql.RuntimeRegisterInput{
		Name:        name,
		Description: &description,
		Labels:      labels,
		Webhooks:    webhooks,
	}
}

func fixModelRuntimeUpdateInput(name, description string) model.RuntimeUpdateInput {
	return model.RuntimeUpdateInput{
		Name:        name,
		Description: &description,
		Labels: map[string]interface{}{
			"test": []string{"val", "val2"},
		},
	}
}

func fixGQLRuntimeUpdateInput(name, description string) graphql.RuntimeUpdateInput {
	labels := graphql.Labels{
		"test": []string{"val", "val2"},
	}

	return graphql.RuntimeUpdateInput{
		Name:        name,
		Description: &description,
		Labels:      labels,
	}
}

func fixModelAuth() *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
		AdditionalHeaders:     map[string][]string{"test": {"foo", "bar"}},
		AdditionalQueryParams: map[string][]string{"test": {"foo", "bar"}},
		RequestAuth: &model.CredentialRequestAuth{
			Csrf: &model.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: "foo.url",
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "boo",
						Password: "far",
					},
				},
				AdditionalHeaders:     map[string][]string{"test": {"foo", "bar"}},
				AdditionalQueryParams: map[string][]string{"test": {"foo", "bar"}},
			},
		},
	}
}

func fixGQLAuth() *graphql.Auth {
	return &graphql.Auth{
		Credential: &graphql.BasicCredentialData{
			Username: "foo",
			Password: "bar",
		},
		AdditionalHeaders:     graphql.HTTPHeaders{"test": {"foo", "bar"}},
		AdditionalQueryParams: graphql.QueryParams{"test": {"foo", "bar"}},
		RequestAuth: &graphql.CredentialRequestAuth{
			Csrf: &graphql.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: "foo.url",
				Credential: &graphql.BasicCredentialData{
					Username: "boo",
					Password: "far",
				},
				AdditionalHeaders:     graphql.HTTPHeaders{"test": {"foo", "bar"}},
				AdditionalQueryParams: graphql.QueryParams{"test": {"foo", "bar"}},
			},
		},
	}
}

func fixModelSystemAuth(id, tenant, runtimeID string, auth *model.Auth) pkgmodel.SystemAuth {
	return pkgmodel.SystemAuth{
		ID:        id,
		TenantID:  &tenant,
		RuntimeID: &runtimeID,
		Value:     auth,
	}
}

func fixGQLSystemAuth(id string, auth *graphql.Auth) *graphql.RuntimeSystemAuth {
	return &graphql.RuntimeSystemAuth{
		ID:   id,
		Auth: auth,
	}
}

func fixModelRuntimeEventingConfiguration(t *testing.T, rawURL string) *model.RuntimeEventingConfiguration {
	validURL := fixValidURL(t, rawURL)
	return &model.RuntimeEventingConfiguration{
		EventingConfiguration: model.EventingConfiguration{
			DefaultURL: validURL,
		},
	}
}

func fixGQLRuntimeEventingConfiguration(url string) *graphql.RuntimeEventingConfiguration {
	return &graphql.RuntimeEventingConfiguration{
		DefaultURL: url,
	}
}

func fixValidURL(t *testing.T, rawURL string) url.URL {
	eventingURL, err := url.Parse(rawURL)
	require.NoError(t, err)
	require.NotNil(t, eventingURL)
	return *eventingURL
}

func fixModelRuntimeContext(id, runtimeID, key, val string) *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        id,
		RuntimeID: runtimeID,
		Key:       key,
		Value:     val,
	}
}

func fixGqlRuntimeContext(id, key, val string) *graphql.RuntimeContext {
	return &graphql.RuntimeContext{
		ID:    id,
		Key:   key,
		Value: val,
	}
}

func fixGQLRtmCtxPage(rtmCtxs []*graphql.RuntimeContext) *graphql.RuntimeContextPage {
	return &graphql.RuntimeContextPage{
		Data: rtmCtxs,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(rtmCtxs),
	}
}

func fixRtmCtxPage(rtmCtxs []*model.RuntimeContext) *model.RuntimeContextPage {
	return &model.RuntimeContextPage{
		Data: rtmCtxs,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(rtmCtxs),
	}
}
