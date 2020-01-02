# Connector authentication with Oathkeeper

## Introduction

Oathkeeper provides several ways to authenticate, authorize, and mutate request.

Appropriate configuration of the mutator of type Hydrator allows to authenticate requests with a one-time token issued by the Connector or to validate the subject of the certificate.


## Oathkeeper configuration

To authenticate requests with a one-time token, configure the Oathkeeper like this: 
```
serve:
  api:
    port: 4456
    host: 127.0.0.1

  proxy:
      port: 4455
      host: 127.0.0.1

access_rules:
  repositories: 
    - file: {PATH_TO_ACCESS_RULES_CONFIG_FILE}

authenticators:
  noop:
    enabled: true

authorizers:
  allow:
    enabled: true

mutators:
  hydrator:
    enabled: true
```

Use this access rules configuration:
```
- id: some-id
  upstream:
    url: http://localhost:3000  # URL of Connector Service GraphQL API
    preserve_host: true
    strip_path: /connector
  match:
    url: http://localhost:4455/connector/graphql  # Oathkeeper proxy URL with path used in Connector
    methods:
      - GET
      - POST
  authenticators:
    - handler: noop
  authorizer:
    handler: allow
  mutators:
    - handler: hydrator
      config: 
        api:
          url: http://localhost:8080/v1/tokens/resolve  # URL of Connector Service REST API authenticator
          retry:
            number_of_retries: 3
            delay_in_milliseconds: 3000
```


## One-time-token authentication

The request to the Connector Service is processed like this:
1. Hydrator mutator calls the `/v1/tokens/resolve` endpoint on the Connector Service Validator API.
2. If the `Connector-Token` header is present, the Connector Service Validator resolves the one-time token and sets the client ID as the value of the `ClientIdFromToken` header.
3. The request goes through the Compass Gateway to the Connector Service.
4. The header `ClientIdFromToken` confirms the request authentication and enables the Connector Service to use the client ID to perform a requested operation.

## Security improvements to implement

To prevent any security issues:
- Strip the `ClientIdFromToken` header either on the Oathkeeper level or in the Connector Service Validator API.

 
