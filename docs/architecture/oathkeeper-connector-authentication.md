# Connector authentication with Oathkeeper

## Introduction

Oathkeeper provides several ways to authenticate, authorize or mutate request.

To Authenticate the request with one time token issued by Connector or to validate certificates subject the mutator of type Hydrator can be used.


## Oathkeeper configuration

One time token authentication can be achieved with the following Oathkeeper configuration:
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

And access rules configuration:
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
            number: 3
            delayInMilliseconds: 3000
```


## One-time-token authentication

The request to Connector Service is processed as follows:
1. Hydrator mutator calls `/v1/tokens/resolve` endpoint on Connector Service Validator API
2. If the `Connector-Token` header is present, the token is resolved and client ID is set as `ClientIdFromToken` header.
3. The request goes through Compass Gateway to Connector Service.
4. If the header `ClientIdFromToken` is present, the client is considered to be authenticated and the client ID is used to perform requested operation.

## Things to consider

To prevent any security issues, the following should be considered:
- `ClientIdFromToken` should be stripped either on Oathkeeper level or in Connector Service Validator API

 
