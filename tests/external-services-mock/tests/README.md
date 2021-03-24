#  External Services Mock integration tests

External Services Mock tests check a contract between Compass and external services. 

The tests cover the following scenarios:
- Audit log scenario
- API specification scenario

## Audit log test scenario

The audit log test performs the following operations:
1. Register an application through the Compass Gateway.
2. Get an audit log from the mock service based on the application's name
3. Compare the audit log with the request for registering the application.

## API specification scenario

The API specification test uses the endpoint that returns a random API specification on every call. It performs the following operations:
1. Register an API Definition with a fetch request.
2. Fetch the API specification.
3. Refetch the API specification and check if it is different from the previous one.
4. Get the API Definition and check if the API specification is equal to the new one.

## Development

To run the test locally, set these environment variables:

| Environment variable   |      Description      |  Default value |
|----------|-------------|:------:|
| **DIRECTOR_URL** |  URL to the Director | `https://compass-gateway.kyma.local/director` |
| **USER_EMAIL** |    Dex static user email   |   `admin@kyma.cx` |
| **USER_PASSWORD** |    Dex static user password   |  None |
| **DEFAULT_TENANT** | Default tenant value |    `3e64ebae-38b5-46a0-b1ed-9ccee153a0ae` |
| **DOMAIN** | Kyma domain name |    `kyma.local` |
| **EXTERNAL_SERVICES_MOCK_BASE_URL** | URL to External Services Mock | None |
| **APP_CLIENT_SECRET**   | The expected audit log client Secret used to obtain a JWT | `client_secret`
| **APP_CLIENT_ID**       | The expected audit log client ID used to obtain a JWT  | `client_id`
| **BASIC_USERNAME**      | The expected username from basic credentials | `admin`
| **BASIC_PASSWORD**      | The expected password from basic credentials | `admin`

After specyfing the environment variables, run `go test ./... -count=1 -v` inside the `./tests/director/external-services-mock-integration` directory.
