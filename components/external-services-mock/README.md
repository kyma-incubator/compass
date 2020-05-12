# External services mock

## Overview

External services mock is a test component that mocks external APIs for test purposes.

Implemented external servies:
* auditlog

## Configuration

External services mock binary allows you to override some configuration parameters. You can specify the following basic environment variables:

| Name                             | Description                                                       | Default value            | 
| ---------------------------------| ----------------------------------------------------------------- | ------------------------ | 
| **APP_ADDRESS**                  | The address and port for the service to listen on                 | `http://127.0.0.1:8080`  | 

### Audit log mock API configuration
| Name                             | Description                                                                       | 
| -------------------------------- | --------------------------------------------------------------------------------- | 
| **APP_AUDITLOG_CLIENT_SECRET**   | The expected audit log client secret which is used in obtaining JWT token         | 
| **APP_AUDITLOG_CLIENT_ID**       | The expected audit log client id which is used in obtaining JWT token              | 

## Development

To run the component locally, export all mandatory and if needed optional ENVs and run:

```bash
make build-local
./bin/external-services-mock
```
