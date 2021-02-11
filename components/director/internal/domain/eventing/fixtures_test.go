package eventing

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

var tenantID = uuid.New()
var externalTenantID = uuid.New()
var runtimeID = uuid.New()
var applicationID = uuid.New()

const (
	eventURLSchema  = "https://eventing.domain.local/%s/v1/events"
	runtimeEventURL = "https://eventing.domain.local"
)

func fixCtxWithTenant() context.Context {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String(), externalTenantID.String())

	return ctx
}

func fixRuntimeEventingURLLabel() *model.Label {
	return &model.Label{
		ID:         uuid.New().String(),
		Key:        RuntimeEventingURLLabel,
		ObjectID:   runtimeID.String(),
		ObjectType: model.RuntimeLabelableObject,
		Tenant:     tenantID.String(),
		Value:      runtimeEventURL,
	}
}

func fixRuntimeEventngCfgWithURL(t *testing.T, rawURL string) *model.RuntimeEventingConfiguration {
	validURL := fixValidURL(t, rawURL)

	return &model.RuntimeEventingConfiguration{
		EventingConfiguration: model.EventingConfiguration{
			DefaultURL: validURL,
		},
	}
}

func fixRuntimeEventngCfgWithEmptyURL(t *testing.T) *model.RuntimeEventingConfiguration {
	return fixRuntimeEventngCfgWithURL(t, EmptyEventingURL)
}

func fixRuntimes() []*model.Runtime {
	return []*model.Runtime{
		&model.Runtime{
			ID:     runtimeID.String(),
			Name:   "runtime-1",
			Tenant: tenantID.String(),
		},
		&model.Runtime{
			ID:     uuid.New().String(),
			Name:   "runtime-2",
			Tenant: tenantID.String(),
		},
	}
}

func fixRuntimePage() *model.RuntimePage {
	modelRuntimes := fixRuntimes()
	return &model.RuntimePage{
		Data:       modelRuntimes,
		TotalCount: len(modelRuntimes),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}
}

func fixRuntimePageWithOne() *model.RuntimePage {
	modelRuntimes := []*model.Runtime{
		fixRuntimes()[0],
	}
	return &model.RuntimePage{
		Data:       modelRuntimes,
		TotalCount: len(modelRuntimes),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}
}

func fixEmptyRuntimePage() *model.RuntimePage {
	return &model.RuntimePage{
		Data:       nil,
		TotalCount: 0,
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}
}

func fixLabelFilterForRuntimeDefaultEventingForApp() []*labelfilter.LabelFilter {
	return []*labelfilter.LabelFilter{
		labelfilter.NewForKey(getDefaultEventingForAppLabelKey(applicationID)),
	}
}

func fixLabelFilterForRuntimeScenarios() []*labelfilter.LabelFilter {
	return []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(model.ScenariosKey, `$[*] ? ( @ == "DEFAULT" || @ == "CUSTOM" )`),
	}
}

func fixApplicationScenariosLabel() *model.Label {
	return &model.Label{
		ID:         uuid.New().String(),
		Key:        model.ScenariosKey,
		ObjectID:   applicationID.String(),
		ObjectType: model.ApplicationLabelableObject,
		Tenant:     tenantID.String(),
		Value:      []interface{}{"DEFAULT", "CUSTOM"},
	}
}

func fixMatcherDefaultEventingForAppLabel() func(l *model.Label) bool {
	return func(l *model.Label) bool {
		return l.Key == getDefaultEventingForAppLabelKey(applicationID)
	}
}

func fixModelApplicationEventingConfiguration(t *testing.T, rawURL string) *model.ApplicationEventingConfiguration {
	validURL := fixValidURL(t, rawURL)
	return &model.ApplicationEventingConfiguration{
		EventingConfiguration: model.EventingConfiguration{
			DefaultURL: validURL,
		},
	}
}

func fixGQLApplicationEventingConfiguration(url string) *graphql.ApplicationEventingConfiguration {
	return &graphql.ApplicationEventingConfiguration{
		DefaultURL: url,
	}
}

func fixValidURL(t *testing.T, rawURL string) url.URL {
	eventingURL, err := url.Parse(rawURL)
	require.NoError(t, err)
	require.NotNil(t, eventingURL)
	return *eventingURL
}

func fixApplicationModel(name string) model.Application {
	return model.Application{
		Tenant:     tenantID.String(),
		Name:       name,
		BaseEntity: &model.BaseEntity{ID: applicationID.String()},
	}
}

func fixAppEventURL(t *testing.T, appName string) url.URL {
	eventURL := fmt.Sprintf(eventURLSchema, appName)
	return fixValidURL(t, eventURL)
}
