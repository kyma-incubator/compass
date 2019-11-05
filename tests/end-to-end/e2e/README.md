# Compass end to end tests

Test Scenario:
- Get Dex token, create IntegrationSystem with it and generate client credentials for IntSystem
- Call Hydra for OAuth 2.0 access token with client_id and client_secret pair - https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials
- Create an Application (as Integration System)
- Add example API Spec using issued OAuth2.0 Access token (as Integration System)
- Remove application as user (using JWT token from Dex)
- Remove IntegrationSystem as user (using JWT token from Dex)
- Test if token granted for IntegrationSystem is invalid
- Test if IntegrationSystem cannot fetch token
