package gqlschema

import (
	"context"

	"github.com/kyma-incubator/compass/components/gateway/internal/director"
	pb "github.com/kyma-incubator/compass/components/gateway/protobuf"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	directorClient *director.Client
}

func NewResolver(directorClient *director.Client) *Resolver {
	return &Resolver{directorClient: directorClient}
}

func (r *Resolver) Application() ApplicationResolver {
	return &applicationResolver{r}
}
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type applicationResolver struct{ *Resolver }

func (r *applicationResolver) Apis(ctx context.Context, obj *Application) ([]*APIDefinition, error) {
	cli, conn, err := r.directorClient.DirectorClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := cli.Apis(ctx, &pb.ApplicationRoot{
		ID: obj.ID,
	})
	if err != nil {
		return nil, err
	}

	var apiDefs []*APIDefinition
	for _, api := range resp.ApplicationApis {
		apiDefs = append(apiDefs, &APIDefinition{
			ID:        api.ID,
			TargetURL: api.TargetURL,
		})
	}

	return apiDefs, nil
}
func (r *applicationResolver) EventAPIs(ctx context.Context, obj *Application) ([]*EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *applicationResolver) Documents(ctx context.Context, obj *Application) ([]*Document, error) {
	panic("not implemented")
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateApplication(ctx context.Context, in ApplicationInput) (*Application, error) {
	panic("not implemented")
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
func (r *mutationResolver) SetAPICredential(ctx context.Context, apiID string, in CredentialInput) (*Credential, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteAPICredential(ctx context.Context, apiID string) (*Credential, error) {
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
	panic("not implemented")
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

func (r *queryResolver) Applications(ctx context.Context, filter []*LabelFilter) ([]*Application, error) {
	cli, conn, err := r.directorClient.DirectorClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	appInput := &pb.ApplicationsInput{}
	resp, err := cli.Applications(ctx, appInput)
	if err != nil {
		return nil, err
	}

	var gqlApps []*Application
	for _, app := range resp.Applications {

		labels := make(map[string]interface{})
		for k, v := range app.Labels {
			labels[k] = v
		}

		annotations := make(map[string]interface{})
		for k, v := range app.Annotations {
			annotations[k] = string(v)
		}
		gqlApps = append(gqlApps, &Application{
			ID:          app.Id,
			Name:        app.Name,
			Tenant:      "string",
			Description: &app.Description,
			Annotations: annotations,
			Labels:      labels,
			Webhooks:    nil,
			Status: &ApplicationStatus{
				//Condition:app.Status.Condition, // TODO:
				Condition: ApplicationStatusConditionReady,
				Timestamp: int(app.Status.Timestamp),
			},
		})
	}

	return gqlApps, nil
}
func (r *queryResolver) Application(ctx context.Context, id string) (*Application, error) {
	panic("not implemented")
}
func (r *queryResolver) Runtimes(ctx context.Context, filter []*LabelFilter) ([]*Runtime, error) {
	panic("not implemented")
}
func (r *queryResolver) Runtime(ctx context.Context, id string) (*Runtime, error) {
	panic("not implemented")
}
func (r *queryResolver) HealthChecks(ctx context.Context, types []HealthCheckType, origin *string) ([]*HealthCheck, error) {
	panic("not implemented")
}
