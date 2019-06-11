package gqlschema

import (
	"context"
	"time"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateApplication(ctx context.Context, in ApplicationInput) (*Application, error) {
	app := &Application{
		ID:     "1",
		Tenant: "abc",
		Status: &ApplicationStatus{
			Condition: ApplicationStatusConditionReady,
			Timestamp: Timestamp(time.Now()),
		},
		Annotations: *in.Annotations,
	}
	return app, nil
}
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, in ApplicationInput) (*Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplication(ctx context.Context, id string) (*Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddApplicationAnnotation(ctx context.Context, applicationID string, annotation string, value string) (string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplicationAnnotation(ctx context.Context, applicationID string, annotation string) (*string, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddApplicationWebhook(ctx context.Context, applicationID string, in ApplicationWebhookInput) (*ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateApplicationWebhook(ctx context.Context, webhookID string, in ApplicationWebhookInput) (*ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplicationWebhook(ctx context.Context, webhookID string) (*ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddAPI(ctx context.Context, applicationID string, in APIDefinitionInput) (*APIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateAPI(ctx context.Context, id string, in APIDefinitionInput) (*APIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteAPI(ctx context.Context, id string) (*APIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) RefetchAPISpec(ctx context.Context, apiID string) (*APISpec, error) {
	panic("not implemented")
}
func (r *mutationResolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in AuthInput) (*RuntimeAuth, error) {
	return &RuntimeAuth{
		RuntimeID: "1",
		Auth: &Auth{
			AdditionalHeaders: in.AdditionalHeaders,
		},
	}, nil
}
func (r *mutationResolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*RuntimeAuth, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddEvent(ctx context.Context, applicationID string, in EventDefinitionInput) (*EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateEvent(ctx context.Context, id string, in EventDefinitionInput) (*EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteEvent(ctx context.Context, id string) (*EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) RefetchEventSpec(ctx context.Context, eventID string) (*EventSpec, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateRuntime(ctx context.Context, in RuntimeInput) (*Runtime, error) {
	return &Runtime{
		ID: "1",
		AgentAuth: &Auth{
			Credential: nil,
			AdditionalQueryParams: &QueryParams{
				"additional": []string{"param1", "param2"},
			},
			AdditionalHeaders: &HttpHeaders{
				"additional": []string{"header1", "header2"},
			},
		},
		Name: in.Name,
	}, nil
}
func (r *mutationResolver) UpdateRuntime(ctx context.Context, id string, in RuntimeInput) (*Runtime, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteRuntime(ctx context.Context, id string) (*Runtime, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddRuntimeLabel(ctx context.Context, runtimeID string, key string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteRuntimeLabel(ctx context.Context, id string, key string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddRuntimeAnnotation(ctx context.Context, runtimeID string, key string, value string) (string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteRuntimeAnnotation(ctx context.Context, id string, key string) (*string, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Applications(ctx context.Context, filter []*LabelFilter, first *int, after *string) (*ApplicationPage, error) {

	app := &Application{
		Tenant: Tenant("abc"),
		Status: &ApplicationStatus{
			Condition: ApplicationStatusConditionReady,
			Timestamp: Timestamp(time.Now()),
		},
		Annotations: Annotations{
			"abc":  "def",
			"abcd": 123,
		},
	}

	return &ApplicationPage{
		Data:       []*Application{app},
		TotalCount: 0,
		PageInfo: &PageInfo{
			HasNextPage: false,
		},
	}, nil
}
func (r *queryResolver) Application(ctx context.Context, id string) (*Application, error) {
	panic("not implemented")
}
func (r *queryResolver) Runtimes(ctx context.Context, filter []*LabelFilter, first *int, after *string) (*RuntimePage, error) {

	rt := &Runtime{
		ID: "1",
		AgentAuth: &Auth{
			Credential: nil,
			AdditionalQueryParams: &QueryParams{
				"additional": []string{"param1", "param2"},
			},
			AdditionalHeaders: &HttpHeaders{
				"additional": []string{"header1", "header2"},
			},
		},
		Name: "name",
	}

	return &RuntimePage{
		Data:       []*Runtime{rt},
		TotalCount: 0,
		PageInfo: &PageInfo{
			HasNextPage: false,
		},
	}, nil
}
func (r *queryResolver) Runtime(ctx context.Context, id string) (*Runtime, error) {
	panic("not implemented")
}
func (r *queryResolver) HealthChecks(ctx context.Context, types []HealthCheckType, origin *string, first *int, after *string) (*HealthCheckPage, error) {
	panic("not implemented")
}
