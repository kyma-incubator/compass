# Gateway

## Overview

Gateway is a component that exposes a service through which it forwards the incoming requests to backing services, such as the Director and Connector. Optionally, Gateway can be configured to send audit logs to the specified logging service.

## Configuration

Gateway binary allows you to override some configuration parameters. You can specify the following basic environment variables:

| Name                             | Default value                                             | Description                                                       | 
| ---------------------------------| --------------------------------------------------------- | ----------------------------------------------------------------- | 
| **APP_ADDRESS**                  | `http://127.0.0.1:3000`                                   | The address and port for the service to listen on                 | 
| **APP_SERVER_TIMEOUT**           | `114s`                                                    | The timeout used for incoming calls to the gateway server         |
| **APP_DIRECTOR_ORIGIN**          | `http://127.0.0.1:3001`                                   | The address and port on which the Director service is listening   | 
| **APP_CONNECTOR_ORIGIN**         | `http://127.0.0.1:3002`                                   | The address and port on which the Connector service is listening  | 
| **APP_AUDITLOG_ENABLED**         | `false`                                                   | The variable that enables the audit log feature                   | 


### Audit log configuration

If you set **APP_AUDITLOG_ENABLED** to `true`, you must specify the following environment variables:

| Name                             | Description                                                                       | 
| -------------------------------- | --------------------------------------------------------------------------------- | 
| **APP_AUDITLOG_URL**             | The URL under which the audit log service is available                            |
| **APP_AUDITLOG_CLIENT_TIMEOUT**  | The timeout used for calls to the audit log service (Default value is `30sec`)    |
| **APP_AUDITLOG_CONFIG_PATH**     | The path for logging configuration changes                                        | 
| **APP_AUDITLOG_SECURITY_PATH**   | The path for logging security events                                              | 
| **APP_AUDITLOG_AUTH_MODE**       | The audit log authorization mode. The possible values are `basic` and `oauth`.    |  
| **APP_AUDITLOG_WRITE_WORKERS**   | The number of goroutines that will consume messages from the channel which will be sent to the Auditlog service (Default value is `5`)| 

Gateway processes audit log messages asynchronously using the configurable Go channel.
The audit log feature reads the messages from the channel and sends them to the audit log service.
You can configure the channel using the following environment variables:

| Name                             | Default value        | Description                                                                       | 
| -------------------------------- | -------------------- | --------------------------------------------------------------------------------- | 
| **APP_AUDITLOG_CHANNEL_SIZE**    |         `100`        | The number of audit log messages that the message channel can store               |  
| **APP_AUDITLOG_CHANNEL_TIMEOUT** |         `5s`         | The time after which sending the message is aborted in case the channel is full   |


If you set **APP_AUDITLOG_AUTH_MODE** to `basic`, you must specify the following environment variables:

| Name                             | Description                                                   |  
| -------------------------------- | ------------------------------------------------------------- |  
| **APP_AUDITLOG_USER**            | The username to the audit log service                         |
| **APP_AUDITLOG_PASSWORD**        | The password to the audit log service                         |
| **APP_AUDITLOG_TENANT**          | The tenant for which audit logs are created                   |

If you set **APP_AUDITLOG_AUTH_MODE** to `oauth`, you must specify the following environment variables:

| Name                              |   Default value  | Required | Description                                                     |
| --------------------------------- | ---------------- |:--------:|---------------------------------------------------------------- |
| **APP_AUDITLOG_CLIENT_ID**        |    None          |   Yes    | The username to the OAuth service                               |
| **APP_AUDITLOG_CLIENT_SECRET**    |    None          |   Yes    | The password to the OAuth service                               |
| **APP_AUDITLOG_OAUTH_URL**        |    None          |   Yes    | The OAuth URL from which Gateway gets the access token          |
| **APP_AUDITLOG_OAUTH_USER**       |   `$USER`        |   No     | The name of the user that is saved in the audit log message     |
| **APP_AUDITLOG_OAUTH_TENANT**     |   `$PROVIDER`    |   No     | The name of the tenant that is saved in the audit log message   |
