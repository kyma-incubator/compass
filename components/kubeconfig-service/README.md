# (OIDC) Kubeconfig Service

## Overview
The Kubeconfig Service is a single purpose [REST](https://en.wikipedia.org/wiki/Representational_state_transfer) based API. It is designed to retrieve a kubeconfig file from an SKR cluster, and change the default authentication mechanism (token), to [kube-login](https://github.com/int128/kubelogin)

## Configuration
The application uses the following Environmental Variables for configuration:

| Parameter | Description | Default value |
| :---: | :--- | :---: | 
| **SERVICE_PORT** | Port used by the application | `8000` |
| **GRAPHQL_URL** | Full URL of the chosen [GraphQL](https://graphql.org/learn/) service | `http://127.0.0.1:3000/graphql` |
| **OIDC_ISSUER_URL** | Full URL of the chosen OIDC Issuer instance | `""` |
| **OIDC_CLIENT_ID** | ClientID for the chosen OIDC Issuer | `""` |
| **OIDC_CLIENT_SECRET** | Client secret for the chosen OIDC Issuer | `""` |

> **NOTE:** All **OIDC** parameters are required in order to start the application, and it is up to the user to provide them.

## HowTo

### Building a local image

Use the provided makefile: 

```bash
make build-image
```

### Run locally

Set your required parameters and run the binary, or use a docker image:

```bash
#Go run
OIDC_ISSUER_URL=https://foo OIDC_CLIENT_ID=foobar OIDC_CLIENT_SECRET=1234 go run cmd/generator/main.go
#Docker run
docker run --rm -e OIDC_ISSUER_URL=https://foo -e OIDC_CLIENT_ID=foobar -e OIDC_CLIENT_SECRET=1234 kubeconfig-service
```

