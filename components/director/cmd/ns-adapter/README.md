# NS-Adapter

## Overview

NS-Adapter listens for incoming **reports** on on-premice systems from external Notification Service. Based on the data from the report systems may be created, updated or deleted.

## Details

This section describes the API schema that a Notification Service must implement to integrate with the NS-Adapter.

### Authorization

Valid JWT token with "tenant.consumerTenant" and "tenant.externalTenant" claims. ConsumerTenant and externalTenant should contain the internal ID and the external ID from existing business_tenant_mappings record in Compass database.

### Reports

There are two types of reports:
- Delta:
  - systems that are __present__ in the report but are __unknown__ to Compass will be created
  - systems that are __present__ in the report and are __known__ to Compass will be updated
  - systems that are __not present__ in the report but are __known__ to Compass will be deleted
- Full:
  - same as Delta
  - systems from Compass labeled with `SCC` are deleted if the value of their label is not present in the report - there is no entity in the `value` array from the request body which has `subaccount` and `locationId` matching these from the systems `SCC` label

|||
|---------------------|--------------------------------------|
|**Description**      | Upsert ExposedSystems is a bulk create-or-update operation on exposed on-premise systems. It takes a list of fully described exposed systems, creates the ones CMP isn't aware of and updates the metadata for the ones it is. |
|**URL**              | /api/v1/notifications |
|                     |                                      |
|**Query Params**     | reportType=full,delta                |
|                     |                                      |
|**HTTP Method**      | PUT                             |
|**HTTP Headers**     |                                      |
|                     | Content-Type: application/json        |
|**HTTP Codes**       |                                      |
|                     | [204 No Content](#204-no-content)    |
|                     | [200 OK](#200-OK)                    |
|                     | [400 Bad Request](#400-bad-request)  |
|                     | [401 Unauthorized](#401-unauthorized)|
|                     | [408 Request Timeout](#408-request-timeout) |
|                     | [500 Internal Server Error](#500-internal-server-error)|
|                     |[502 Bad Gateway](#502-bad-gateway)|
|                     |[413 Payload Too Large](#413-payload-too-large)|
|**Response Formats** | json                                 |
|**Authentication**   | OAuth 2 via XSUAA |


Example request with delta report:
```
curl -v --request PUT \
  --url http://{domain}/api/v1/notifications/?reportType=delta \
  --header 'content-type: application/json' \
  --data \
'
{
  "type": "notification-service",
  "value": [
    {
      "subaccount": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "locationId": "location-id-1",
      "exposedSystems": [
        {
          "protocol": "HTTP",
          "host": "127.0.0.1:8080",
          "type": "on-premice-system",
          "status": "disabled",
          "description": "system description",
          "systemNumber": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
        }
      ]
    },
     {
      "subaccount": "cccccccc-cccc-cccc-cccc-cccccccccccc"
      "locationId": "location-id-2",
      "exposedSystems": [
        {
          "protocol": "HTTP",
          "host": "127.0.0.1:8080",
          "type": "on-premice-system",
          "status": "disabled",
          "description": "system description",
          "systemNumber": "dddddddd-dddd-dddd-dddd-dddddddddddd"
        }
      ]
    }
  ]
}
'
```

Example request with full report:
```
curl -v --request PUT \
  --url http://{domain}/api/v1/notifications/?reportType=full \
  --header 'content-type: application/json' \
  --data \
'
{
  "type": "notification-service",
  "value": [
    {
      "subaccount": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "locationId": "location-id-1",
      "exposedSystems": [
        {
          "protocol": "HTTP",
          "host": "127.0.0.1:8080",
          "type": "on-premice-system",
          "status": "disabled",
          "description": "system description",
          "systemNumber": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
        }
      ]
    },
     {
      "subaccount": "cccccccc-cccc-cccc-cccc-cccccccccccc"
      "locationId": "location-id-2",
      "exposedSystems": [
        {
          "protocol": "HTTP",
          "host": "127.0.0.1:8080",
          "type": "on-premice-system",
          "status": "disabled",
          "description": "system description",
          "systemNumber": "dddddddd-dddd-dddd-dddd-dddddddddddd"
        }
      ]
    }
  ]
}
'
```

### NS-Adapter Response


#### 204 No Content

Successfully created or updated all the exposed systems in the list. In case of a full report, this is the only possible response which might be returned - it will not return list of SCCs for which the update failed.
Almost every top-level data that is expected by the tenant fetcher is configurable. Consider the following example response for the creation endpoint:

```
HTTP/1.1 204 No Content
```

#### 200 OK

If upsert for one or more on-premise systems has failed. This can be returned only by delta reports, but not by a full report as the full report will not return SCCs for which the update failed.

```
HTTP/1.1 200 OK
{
  "error": {
    "code": "200",
    "message": "Update/create failed for some on-premise systems",
    "details": [{
      "message": "Creation failed",
      "subaccount": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "locationId": "loc-id1"
    }]
  }
}
```

#### 400 Bad Request

##### Failed Deserialization

If the service cannot deserialize the request.

```
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": 400,
    "message": "failed to parse request body"
  }
}
```

##### Failed to read request body

If the service cannot read the request body from the stream.

```
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": 400,
    "message": "faild to retrieve request body"
  }
}
```

##### Missing Required Property

In case any of the required properties is not provided, e.g. subaccount.

```
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": 400,
    "message": "subaccount key is required"
  }
}
```

##### Missing query parameter
In case the mandatory query parameter reportType is missing
```
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": 400,
    "message": "the query parameter 'reportType' is missing or invalid"
  }
}
```
#### 401 Unauthorized

In the event of a failed authentication.

```
HTTP/1.1 401 Unauthoriezd
Content-Type: application/json

{
  "error": {
    "code": 401,
    "message": "unauthorized"
  }
}
```
#### 408 Request Timeout

In the event of request timeout in CMP infrastructure
```
HTTP/1.1 408 Request Timeout
Content-Type: application/json

{
  "error": {
    "code": 408,
    "message": "timeout"
  }
}
```
#### 500 Internal Server Error

In the event of any internal failure during processing.

```
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": {
    "code": 500,
    "message": "Update failed"
  }
}
```
#### 502 Bad Gateway

In the event of communication error in CMP infrastructure
```
HTTP/1.1 502 Bad Gateway
Content-Type: text/plain

<content>
```
#### 413 Payload Too Large

In the event of request body larger than 5 MB
```
HTTP/1.1 413 Payload Too Large
Content-Type: text/plain

<content>
```

## Configuration

NS-Adapter binary allows you to override some configuration parameters. Up-to-date list of the configurable parameters of NS-Adapter can be found [here](https://github.com/kyma-incubator/compass/blob/66754e58f6d23f233a6905fa9e4577b6666c98e7/components/director/internal/nsadapter/adapter/configuration.go#L13)

## Local Development

### Prerequisites

NS-Adapter requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
2. Service that exposes endpoint for fetching JWKSs

### Run
You can start the NS-Adapter via the designated `run.sh` script in Director:
```bash
./run.sh --ns-adapter --jwks-endpoint {your_JWKS_endpoint}

```

That will bring up a PostgreSQL database with Director's schema.

The up-to-date properties supported by the `run.sh` script can be found in the [Director component local documentation](../../README.md#local-development).