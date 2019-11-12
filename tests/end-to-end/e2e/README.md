# Compass end to end tests

## Test Scenario:
- Get Dex token, create IntegrationSystem with it and generate client credentials for IntSystem
- Call Hydra for OAuth 2.0 access token with client_id and client_secret pair - https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials
- Create an Application (as Integration System)
- Add example API Spec using issued OAuth2.0 Access token (as Integration System)
- Remove application as user (using JWT token from Dex)
- Remove IntegrationSystem as user (using JWT token from Dex)
- Test if token granted for IntegrationSystem is invalid
- Test if IntegrationSystem cannot fetch token

## How to run test locally
To run the test locally, set these environment variables:

| Name   |      Description      |  Default value |
|----------|:-------------:|------:|
| DIRECTOR_URL |  URL to Compass Director | `https://compass-gateway.kyma.local/director` |
| USER_EMAIL |    Dex static user email   |   `admin@kyma.cx` |
| USER_PASSWORD |    Dex static user password   |   - |
| DEFAULT_TENANT | Default tenant value |    `3e64ebae-38b5-46a0-b1ed-9ccee153a0ae` |
| DOMAIN | Kyma domain name |    `kyma.local` |

Then run `go test ./... -count=1 -v` inside `./tests/end-to-end/e2e` directory.
