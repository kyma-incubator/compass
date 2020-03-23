# Gateway

## Overview

Gateway is a component that exposes a service through which it forwards the incoming requests to backing services, such as the Director and Connector. Optionally, Gateway can be configured to send audit logs to the specified logging service.

## Configuration

Gateway binary allows you to override some configuration parameters. You can specify the following basic environment variables:

| Name                             | Description                                                     | Default value            | 
| ---------------------------------| ----------------------------------------------------------------- | ------------------------ | 
| **APP_ADDRESS**                  | The address and port for the service to listen on                 | `http://127.0.0.1:3000`  | 
| **APP_DIRECTOR_ORIGIN**          | The address and port on which the Director service is listening   | `http://127.0.0.1:3001`  | 
| **APP_CONNECTOR_ORIGIN**         | The address and port on which the Connector service is listening  | `http://127.0.0.1:3002`  | 
| **APP_AUDITLOG_ENABLED**         | The variable that enables the audit log feature                   | `false`                  | 

### Audit log authorization configuration

If you set **APP_AUDITLOG_ENABLED** to `true`, the following environment variables are required:

| Name                             | Description                                                                     | 
| -------------------------------- | ------------------------------------------------------------------------------- |
| **APP_AUDITLOG_URL**             | The URL under which the audit log service is available                          |
| **APP_AUDITLOG_CONFIG_PATH**     | The path for logging configuration changes                                      |
| **APP_AUDITLOG_SECURITY_PATH**   | The path for logging security events                                            |
| **APP_AUDITLOG_AUTH_MODE**       | The audit log authorization mode. The possible values are `basic` and `oauth`.  |

If you set **APP_AUDITLOG_AUTH_MODE** to `basic`, you are required to pass the following values:

| Name                             | Description                                                    |  
| -------------------------------- | -------------------------------------------------------------- |  
| **APP_AUDITLOG_USER**            | The username to the audit log service                          |
| **APP_AUDITLOG_PASSWORD**        | The password to the audit log service                          |
| **APP_AUDITLOG_TENANT**          | The tenant for which audit logs are created                    |

If you set **APP_AUDITLOG_AUTH_MODE** to `oauth`, you are required to pass the following values:

| Name                              | Description                                                |  
| --------------------------------- | ---------------------------------------------------------- |  
| **APP_AUDITLOG_CLIENT_ID**        | The username to the OAuth service service                  |
| **APP_AUDITLOG_CLIENT_SECRET**    | The password to the OAuth service                          |
| **APP_AUDITLOG_OAUTH_URL**        | The OAuth URL from which Gateway gets the access token     |
