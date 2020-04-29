#  External Services Mock integration tests

External Services Mock tests check a contract between Compass and external services. 

## Auditlog test scenario
- Register an application through Compass Gateway
- Get an auditlog from the mock service based on the application's name
- Compare the auditlog with the request for registering the application

## Development
To run the test locally, set these environment variables:

| Environment variable   |      Description      |  Default value |
|----------|:-------------:|------:|
| **DIRECTOR_URL** |  URL to the Director | `https://compass-gateway.kyma.local/director` |
| **USER_EMAIL** |    Dex static user email   |   `admin@kyma.cx` |
| **USER_PASSWORD** |    Dex static user password   |   - |
| **DEFAULT_TENANT** | Default tenant value |    `3e64ebae-38b5-46a0-b1ed-9ccee153a0ae` |
| **DOMAIN** | Kyma domain name |    `kyma.local` |
| **AUDITLOG_MOCK_BASE_URL** | URL to the auditlog service | - |

After specyfing the environment variables, run `go test ./... -count=1 -v` inside the `./tests/director/external-services-mock-integration` directory.
