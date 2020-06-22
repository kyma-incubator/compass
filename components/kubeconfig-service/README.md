# (OIDC) Kubeconfig Service

## Overview

The Kubeconfig Service is a single purpose [REST](https://en.wikipedia.org/wiki/Representational_state_transfer)-based API. It is designed to retrieve a `kubeconfig` file from an SKR cluster and change the default authentication mechanism (token) to [kubelogin](https://github.com/int128/kubelogin).

## Configuration

The application uses the following environment variables for configuration:

| Parameter | Required | Description | Default value |
| :---: | :--- | :--- | :---: | 
| **PORT_SERVICE** | No | Port used by the application. | `8000` |
| **PORT_HEALTH** | No | Port used by the application for health check. | `9000` |
| **GRAPHQL_URL** | Yes | Full URL of the chosen [GraphQL](https://graphql.org/learn/) service. | `http://127.0.0.1:3000/graphql` |
| **OIDC_KUBECONFIG_ISSUER_URL** | Yes | Full URL of the chosen OIDC Issuer instance used for the `kubeconfig` generation. | None |
| **OIDC_KUBECONFIG_CLIENT_ID** | Yes | ClientID for the chosen OIDC Issuer used for the `kubeconfig` generation. | None |
| **OIDC_KUBECONFIG_CLIENT_SECRET** | Yes | Client Secret for the chosen OIDC Issuer used for the `kubeconfig` generation. | None |
| **OIDC_ISSUER_URL** | Yes | Full URL of the chosen OIDC Issuer instance. | None |
| **OIDC_CLIENT_ID** | Yes | ClientID for the chosen OIDC Issuer. | `""` |
| **OIDC_CA** | No | CA certificate file path. | None |
| **OIDC_CLAIM_USERNAME** | No | Identifier of the user in JWT claims. | `email` |
| **OIDC_CLAIM_GROUPS** | No | Identifier of groups in JWT claims. | `groups` |
| **OIDC_USERNAME_PREFIX** | No | If provided, all users are prefixed with this value to prevent conflicts with other authentication strategies. | None |
| **OIDC_GROUPS_PREFIX** | No | If provided, all groups are prefixed with this value to prevent conflicts with other authentication strategies. | None |
| **OIDC_SUPPORTED_SIGNING_ALGS** | No | List of supported signing algorithms. | `RS256` |

## Usage

### Build a local image

To build a local image, use the provided makefile: 

```bash
make build-image
```

### Run locally

Set the required parameters and run the binary, or use a Docker image:

```bash
#Go run
OIDC_KUBECONFIG_ISSUER_URL=https://foo OIDC_KUBECONFIG_CLIENT_ID=foobar OIDC_KUBECONFIG_CLIENT_SECRET=1234 OIDC_ISSUER_URL=https://dex.kyma.local OIDC_CLIENT_ID=compass-ui OIDC_CA=~/.minikube/ca.crt go run cmd/generator/main.go

#Docker run
docker run --rm -e OIDC_KUBECONFIG_ISSUER_URL=https://foo -e OIDC_KUBECONFIG_CLIENT_ID=foobar -e OIDC_KUBECONFIG_CLIENT_SECRET=1234 -e OIDC_ISSUER_URL=https://dex.kyma.local -e OIDC_CLIENT_ID=compass-ui -e OIDC_CA=~/.minikube/ca.crt kubeconfig-service
```
