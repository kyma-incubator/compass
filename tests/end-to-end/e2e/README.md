# Compass end to end tests

Compass end to end test scenario consists of the following steps:
- Get a Dex token, register the Integration System using this token, and request client credentials for the IntegrationSystem
- Call Hydra for OAuth 2.0 access token with client_id and client_secret pair - https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials
- Register an Application as the Integration System
- Add example API Spec using issued OAuth2.0 Access token (as Integration System)
- Try removing the Integration System
- Remove Application using issued OAuth2.0 Access token (test if the token is still valid)
- Remove IntegrationSystem as user (using JWT token from Dex)
- Test if token granted for Integration System is invalid
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
| GATEWAY_JWTSUBDOMAIN | Default gateway for handling requests with a JWT | compass-gateway |
| GATEWAY_CLIENT_CERTS_SUBDOMAIN | Default gateway for handling requests with a certificate | compass-gateway-mtls |
| GATEWAY_OAUTH20_SUBDOMAIN | Default gateway for handling requests with an OAuth access token | compass-gateway-auth-oauth|
Then run `go test ./... -count=1 -v` inside `./tests/end-to-end/e2e` directory.
