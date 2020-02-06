# Kyma Environment Broker

## Overview

Kyma Environment Broker (KEB) is a component that allows you to provision Kyma as a Runtime on clusters provided by third-party providers. It uses Provisioner's API to install Kyma on a given cluster.

For more information, read the [documentation](../../docs/kyma-environment-broker).


## Configuration

KEB binary allows you to override some configuration parameters. You can specify the following environment variables:

| Name | Description | Default value |
|-----|---------|:--------:|
| **APP_PORT** | Specifies the port on which the HTTP server listens. | `8080` |
| **APP_PROVISIONING_URL** | Specifies a URL to the Provisioner's API. | None |
| **APP_PROVISIONING_SECRET_NAME** | Specifies the name of the Secret which holds credentials to the Provisioner's API. | None |
| **APP_PROVISIONING_GARDENER_PROJECT_NAME** | Defines the Gardener project name. | `true` |
| **APP_PROVISIONING_GCP_SECRET_NAME** | Defines the name of the Secret which holds credentials to GCP. | None |
| **APP_PROVISIONING_AWS_SECRET_NAME** | Defines the name of the Secret which holds credentials to AWS. | None |
| **APP_PROVISIONING_AZURE_SECRET_NAME** | Defines the name of the Secret which holds credentials to Azure. | None |
| **APP_AUTH_USERNAME** | Specifies the Kyma Environment Service Broker authentication username. | None |
| **APP_AUTH_PASSWORD** | Specifies the Kyma Environment Service Broker authentication password. | None |
| **APP_DIRECTOR_NAMESPACE** | Specifies the Namespace in which Director is deployed. | `compass-system` |
| **APP_DIRECTOR_URL** | Specifies the Director's URL. | `http://compass-director.compass-system.svc.cluster.local:3000/graphql` |
| **APP_DIRECTOR_OAUTH_CREDENTIALS_SECRET_NAME** | Specifies the name of the Secret created by the Integration System. | `compass-kyma-environment-broker-credentials` |
| **APP_DIRECTOR_SKIP_CERT_VERIFICATION** | Specifies whether TLS checks the presented certificates. | `false` |
| **APP_DATABASE_USER** | Defines the database username. | `postgres` |
| **APP_DATABASE_PASSWORD** | Defines the database user password. | `password` |
| **APP_DATABASE_HOST** | Defines the database host. | `localhost` |
| **APP_DATABASE_PORT** | Defines the database port. | `5432` |
| **APP_DATABASE_NAME** | Defines the database name. | `broker` |
| **APP_DATABASE_SSL** | Specifies the SSL Mode for PostgrSQL. See all the possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html).  | `disable`|


## Development

This section presents how to add or remove the possibility to disable components, which makes them either optional or required during the Kyma installation.

### Add option to disable components

If you want to add the possibility to disable components and make them optional during Kyma installation, you can do it in two ways.

* If disabling a given component only means to remove it from the list, use the generic disabler:

```go
runtime.NewGenericComponentDisabler("component-name", "component-namespace")
``` 

* If disabling a given component requires more complex logic, create a new file called `internal/runtime/{compoent-name}_disabler.go` and implement a service which fulfills the following interface:

```go
// OptionalComponentDisabler disables component form the given list and returns a modified list
type OptionalComponentDisabler interface {
	Disable(components internal.ComponentConfigurationInputList) internal.ComponentConfigurationInputList
```

>**NOTE**: Check the [LokiDisabler](`internal/runtime/loki_disabler.go`) as an example of custom service for disabling components.

In each method, the framework injects the  **components** parameter which is a list of components that are sent to the Runtime Provisioner. The implemented method is responsible for disabling component and as a result, returns a modified list. 
  
This interface allows you to easily register the disabler in the [`cmd/broker/main.go`](./cmd/broker/main.go) file by adding a new entry in the **optionalComponentsDisablers** list:

```go
// Register disabler. Convention:
// {component-name} : {component-disabler-service}
//
// Using map is intentional - we ensure that component name is not duplicated.
optionalComponentsDisablers := runtime.ComponentsDisablers{
		"Loki":       runtime.NewLokiDisabler(),
		"Kiali":      runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger":     runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
		"Monitoring": runtime.NewGenericComponentDisabler("monitoring", "kyma-system"),
}
```

### Remove option to disable components

If you want to remove the option to disable components and make them required during Kyma installation, remove a given entry from the **optionalComponentsDisablers** list in the [`cmd/broker/main.go`](./cmd/broker/main.go) file.
