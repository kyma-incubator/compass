# External Services Mock

## Overview

External Services Mock is a component that mocks external APIs for test purposes. To learn more about tests that use the External Services Mock, read [this](https://github.com/kyma-incubator/compass/blob/main/tests/external-services-mock/tests/README.md) document.

External Services Mock contains the following mocks:
* Audit log service
* API specification service

## Configuration

External Services Mock binary allows you to override some configuration parameters. You can specify the following basic environment variables:

| Name                             | Description                                                       | Default value            | 
| ---------------------------------| ----------------------------------------------------------------- | ------------------------ | 
| **APP_ADDRESS**                  | The address and port for the service to listen on                 | `http://127.0.0.1:8080`  | 

### Audit log mock API configuration
| Name                             | Description                                                                       | 
| -------------------------------- | --------------------------------------------------------------------------------- | 
| **APP_CLIENT_SECRET**   | The expected audit log client Secret used to obtain a JWT         | 
| **APP_CLIENT_ID**       | The expected audit log client ID used to obtain a JWT             | 
| **BASIC_USERNAME**      | The expected username from basic credentials                      |
| **BASIC_PASSWORD**      | The expected password from basic credentials                      |
## Development

Use this command to run the component locally:

```bash
export APP_CLIENT_SECRET={CLIENT_SECRET}
export APP_CLIENT_ID={CLIENT_ID}
export BASIC_USERNAME={USERNAME}
export BASIC_PASSWORD={PASSWORD}
go run ./cmd/main.go
```
