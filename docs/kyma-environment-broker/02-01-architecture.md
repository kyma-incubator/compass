# Architecture

The diagram and steps describe the Kyma Environment Broker (KEB) workflow and the roles of specific  components in this process:

![KEB diagram](./assets/keb-architecture.svg)

1. The user sends a request to create a new cluster with Kyma Runtime.
2. KEB proxies the request to create a new cluster to the Provisioner component.
3. Provisioner registers a new cluster in the Director component.
4. Provisioner creates a new cluster and installs Kyma Runtime.
5. KEB sends a request to Hydra to refresh the OAuth token.
6. KEB passes the OAuth token through Gateway to Director.
7. Director returns the Dashboard URL, which is the URL to the newly created cluster, through Gateway to KEB.
