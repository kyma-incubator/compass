# Gateway

## Overview

Gateway is a component that exposes a service through which it forwards the incoming requests to backing services (Director and Connector). Optionally, Gateway can be configured to send audit logs to specified logging service.

## Configuration

Gateway binary allows you to override some configuration parameters. You can specify the following basic environment variables:

| Name                             | Description                                                     | Default value            | 
| ---------------------------------| ----------------------------------------------------------------- | ------------------------ | 
| **APP_ADDRESS**                  | The address and port for the service to listen on                 | `http://127.0.0.1:3001`  | 
| **APP_DIRECTOR_ORIGIN**          | The address and port on which the Director service is listening   | `http://127.0.0.1:3000`  | 
| **APP_CONNECTOR_ORIGIN**         | The address and port on which the Connector service is listening  | `http://127.0.0.1:3000`  | 
| **APP_AUDITLOG_ENABLED**         | The variable that enables the audit log feature                   | `false`                  | 

If you set **APP_AUDITLOG_ENABLED** to `true`, the following environment variables are required:

| Name                             | Description                                                    | 
| -------------------------------- | -------------------------------------------------------------- |
| **APP_AUDITLOG_USER**            | The username to the audit log service                          |
| **APP_AUDITLOG_PASSWORD**        | The password to the audit log service                          |
| **APP_AUDITLOG_URL**             | The URL under which the audit log service is available         |
| **APP_AUDITLOG_TENANT**          | The tenant for whom the audit logs are created                 |
| **APP_AUDITLOG_CONFIG_PATH**     | The path for logging configuration changes events              |
| **APP_AUDITLOG_SECURITY_PATH**   | The path for logging security events                           |
