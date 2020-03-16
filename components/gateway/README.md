# Gateway

## Overview

Gateway is a component that aggregates Director's and Connector's GraphQL schemas under one service and then sends proper requests to the Director or Connector accordingly.

## Configuration

Gateway binary allows you to override some configuration parameters. You can specify the following basic environment variables:

| Name                             | Description                                                  | Default                  | 
| ---------------------------------| ------------------------------------------------------------ | ------------------------ | 
| **APP_ADDRESS**                  | The address and port for the service to listen on            | `http://127.0.0.1:3001`  | 
| **APP_DIRECTOR_ORIGIN**          | The address and port for the Director service to listen on   | `http://127.0.0.1:3000`  | 
| **APP_CONNECTOR_ORIGIN**         | The address and port for the Connector service to listen on  | `http://127.0.0.1:3000`  | 
| **APP_AUDITLOG_ENABLED**         | The variable that enables the audit log feature              | `false`                  | 

If you set **APP_AUDITLOG_ENABLED** to `true`, the following environment variables are required:

| Name                             | Description                                                    | Default           | 
| -------------------------------- | -------------------------------------------------------------- | ----------------- |
| **APP_AUDITLOG_USER**            | The username to the audit log service                          |                   |
| **APP_AUDITLOG_PASSWORD**        | The password to the audit log service                          |                   |
| **APP_AUDITLOG_URL**             | The URL under which the audit log service is available         |                   |
| **APP_AUDITLOG_TENANT**          | The tenant for whom the audit logs are logged                  |                   |
| **APP_AUDITLOG_CONFIG_PATH**     | The path for logging configuration changes logs                |                   |
| **APP_AUDITLOG_SECURITY_PATH**   | The path for logging security events                            |                   |
