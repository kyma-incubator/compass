#  Gateway integration tests

Gateway integration tests are a collection of the following tests:
- Playground tests check if the playground is available through Gateway
- Viewer tests check the correctness of the `viewer` mutation
- Tenant Mapping Handler checks tenant separation
- Compass Authentication checks different methods of access to Director via Gateway

## Compass authentication test scenario
Director authentication test scenario consists of the following steps:
- Get Compass's externally-issued client certificate, register the Integration System using the certificate, and request client credentials for the IntegrationSystem
- Call Hydra for OAuth 2.0 access token with client_id and client_secret pair - https://github.com/kyma-incubator/examples/tree/main/ory-hydra/scenarios/client-credentials
- Register an Application as the Integration System
- Add example API Spec using issued OAuth2.0 Access token (as Integration System)
- Try removing the Integration System
- Remove Application using issued OAuth2.0 Access token (test if the token is still valid)
- Remove IntegrationSystem as user (using Compass's externally-issued client certificate)
- Test if token granted for Integration System is invalid
- Test if Integration System cannot fetch token
