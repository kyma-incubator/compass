# Gateway

## Overview

Gateway is a component that exposes a service through which it forwards the incoming requests to backing services, such as the Director and Connector. Optionally, Gateway can be configured to send audit logs to the specified logging service.

## Configuration

Gateway binary allows you to override some configuration parameters. You can specify the following basic environment variables:

| Name                             | Description                                                       | Default value            | 
| ---------------------------------| ----------------------------------------------------------------- | ------------------------ | 
| **APP_ADDRESS**                  | The address and port for the service to listen on                 | `http://127.0.0.1:3000`  | 
| **APP_DIRECTOR_ORIGIN**          | The address and port on which the Director service is listening   | `http://127.0.0.1:3001`  | 
| **APP_CONNECTOR_ORIGIN**         | The address and port on which the Connector service is listening  | `http://127.0.0.1:3002`  | 
| **APP_AUDITLOG_ENABLED**         | The variable that enables the audit log feature                   | `false`                  | 

### Audit log configuration

If you set **APP_AUDITLOG_ENABLED** to `true`, you must specify the following environment variables:

| Name                             | Description                                                                       |   Default value  | 
| -------------------------------- | --------------------------------------------------------------------------------- | ---------------- | 
| **APP_AUDITLOG_URL**             | The URL under which the audit log service is available                            |      None        | 
| **APP_AUDITLOG_CONFIG_PATH**     | The path for logging configuration changes                                        |      None        | 
| **APP_AUDITLOG_SECURITY_PATH**   | The path for logging security events                                              |      None        | 
| **APP_AUDITLOG_AUTH_MODE**       | The audit log authorization mode. The possible values are `basic` and `oauth`.    |      None        |  

Gateway process AuditLog messages asynchronously using configurable Go channel.
The audit log feature reads the messages from the channel and sends them to the audit log service.
You can configure the channel using the following environment variables:

| Name                             | Description                                                                       |   Default value  | 
| -------------------------------- | --------------------------------------------------------------------------------- | ---------------- | 
| **APP_AUDITLOG_CHANNEL_SIZE**    | The number of audit log messages that the message channel can store               |     `100`        |  
| **APP_AUDITLOG_CHANNEL_TIMEOUT** | The time after which sending the message is aborted in case the channel is full   |     `5s`         |


If you set **APP_AUDITLOG_AUTH_MODE** to `basic`, you must specify the following environment variables:

| Name                             | Description                                                   |  
| -------------------------------- | ------------------------------------------------------------- |  
| **APP_AUDITLOG_USER**            | The username to the audit log service                         |
| **APP_AUDITLOG_PASSWORD**        | The password to the audit log service                         |
| **APP_AUDITLOG_TENANT**          | The tenant for which audit logs are created                   |

If you set **APP_AUDITLOG_AUTH_MODE** to `oauth`, you must specify the following environment variables:

| Name                              | Description                                                     |   Default value  | Required|
| --------------------------------- | --------------------------------------------------------------- |----------------- |---------|
| **APP_AUDITLOG_CLIENT_ID**        | The username to the OAuth service                               |    None          |   Yes   |
| **APP_AUDITLOG_CLIENT_SECRET**    | The password to the OAuth service                               |    None          |   Yes   |
| **APP_AUDITLOG_OAUTH_URL**        | The OAuth URL from which Gateway gets the access token          |    None          |   Yes   |
| **APP_AUDITLOG_OAUTH_USER**       | The name of the user that is saved in the audit log message     |   `$USER`        |   No    |
| **APP_AUDITLOG_OAUTH_TENANT**     | The name of the tenant that is saved in the audit log message   |   `$PROVIDER`    |   No    |
