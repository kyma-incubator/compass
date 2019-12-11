package director

import (
	gql "github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Service
type Service interface {
	CreateRuntime(config *gqlschema.RuntimeInput) (string, error)
	DeleteRuntime() error
	UpdateRuntime() error
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer
	runtimeConfig string
}

func NewDirectorClient(gqlClient gql.Client) Service {
	return &directorClient{
		gqlClient:     gqlClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer{},
	}
}

func (cc *directorClient) CreateRuntime(config *gqlschema.RuntimeInput) (string, error) {

	if config == nil {
		return "", errors.New("Cannot register register runtime in Director: missing Runtime config")
	}

	response := CreateRuntimeResponse{}

	graphQLized, err := cc.graphqlizer.RuntimeInputToGraphQL(*config)

	if err != nil {
		return "", err
	}

	applicationsQuery := cc.queryProvider.createRuntimeMutation(graphQLized)
	req := gcli.NewRequest(applicationsQuery)
	//req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant) ???

	err = cc.gqlClient.Do(req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to register runtime in Director")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", errors.Errorf("Failed to register runtime in Director: received nil response.")
	}

	return "", nil
}

func (cc *directorClient) DeleteRuntime() error {
	return nil
}

func (cc *directorClient) UpdateRuntime() error {
	return nil
}

/*
package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"kyma-project.io/compass-runtime-agent/internal/config"
	gql "kyma-project.io/compass-runtime-agent/internal/graphql"
	kymamodel "kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

const (
	TenantHeader = "Tenant"

	eventsURLLabelKey  = "runtime/event_service_url"
	consoleURLLabelKey = "runtime/console_url"
)

type RuntimeURLsConfig struct {
	EventsURL  string `envconfig:"default=https://gateway.kyma.local"`
	ConsoleURL string `envconfig:"default=https://console.kyma.local"`
}

//go:generate mockery -name=DirectorClient
type DirectorClient interface {
	FetchConfiguration() ([]kymamodel.Application, error)
	SetURLsLabels(urlsCfg RuntimeURLsConfig) (graphql.Labels, error)
}

func NewConfigurationClient(gqlClient gql.Client, runtimeConfig config.RuntimeConfig) DirectorClient {
	return &directorClient{
		gqlClient:     gqlClient,
		queryProvider: queryProvider{},
		runtimeConfig: runtimeConfig,
	}
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	runtimeConfig config.RuntimeConfig
}

func (cc *directorClient) FetchConfiguration() ([]kymamodel.Application, error) {
	response := ApplicationsForRuntimeResponse{}

	applicationsQuery := cc.queryProvider.applicationsForRuntimeQuery(cc.runtimeConfig.RuntimeId)
	req := gcli.NewRequest(applicationsQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch Applications")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return nil, errors.Errorf("Failed fetch Applications for Runtime from Director: received nil response.")
	}

	// TODO: After implementation of paging modify the fetching logic

	applications := make([]kymamodel.Application, len(response.Result.Data))
	for i, app := range response.Result.Data {
		applications[i] = app.ToApplication()
	}

	return applications, nil
}

func (cc *directorClient) SetURLsLabels(urlsCfg RuntimeURLsConfig) (graphql.Labels, error) {
	eventsURLLabel, err := cc.setURLLabel(eventsURLLabelKey, urlsCfg.EventsURL)
	if err != nil {
		return nil, err
	}

	consoleURLLabel, err := cc.setURLLabel(consoleURLLabelKey, urlsCfg.ConsoleURL)
	if err != nil {
		return nil, err
	}

	return graphql.Labels{
		eventsURLLabel.Key:  eventsURLLabel.Value,
		consoleURLLabel.Key: consoleURLLabel.Value,
	}, nil
}

func (cc *directorClient) setURLLabel(key, value string) (*graphql.Label, error) {
	response := SetRuntimeLabelResponse{}

	setLabelQuery := cc.queryProvider.setRuntimeLabelMutation(cc.runtimeConfig.RuntimeId, key, value)
	req := gcli.NewRequest(setLabelQuery)
	req.Header.Set(TenantHeader, cc.runtimeConfig.Tenant)

	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to set %s Runtime label to value %s", key, value)
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return nil, errors.Errorf("Failed to set %s Runtime label to value %s. Received nil response.", key, value)
	}

	return response.Result, nil
}

*/
