# Gateway

## Configuration

The Gateway binary allows to override some configuration parameters. You can specify following basic environment variables.

| ENV                                      | Default                         | Description                                                  |
| ---------------------------------------- | ------------------------------- | ------------------------------------------------------------ |
| APP_ADDRESS                              | http://127.0.0.1:3001           | The address and port for the service to listen on            |
| APP_DIRECTOR_ORIGIN                      | http://127.0.0.1:3000           | The address and port for the director service to listen on   |
| APP_CONNECTOR_ORIGIN                     | http://127.0.0.1:3000           | The address and port for the director service to listen on   |
| APP_AUDITLOG_ENABLED                     | false                           | Auditlog feature                                             |

If `APP_AUDITLOG_ENABLED` is set to true, the following envs are required:

| ENV                                      | Description                                                    |
| ---------------------------------------- | -------------------------------------------------------------- |
| APP_AUDITLOG_USER                        | The user which is used to authenticate to auditlog service     |     
| APP_AUDITLOG_PASSWORD                    | The password which is used to authenticate to auditlog service |
| APP_AUDITLOG_URL                         | The url where auditlog service is availables                   |
| APP_AUDITLOG_TENANT                      | Tenant for whom the auditlogs will be logged                   |
| APP_AUDITLOG_CONFIG_PATH                 | The path for logging configuration changes logs                |
| APP_AUDITLOG_SECURITY_PATH               | The path for loggin security events                            |
