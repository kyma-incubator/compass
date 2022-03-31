# Notification Service Adapter (NS-Adapter)

## Overview of terms

- *SAP Cloud Connector (SCC)* - A component, which runs on-premise in a customer data center. It provides connectivity between Cloud and on-premise systems.
- *Notification Service (NS)* - A Cloud service where various SCCs are registered. Cloud applications connect to on-premise systems, which are exposed by these registered SCCs. An SCC registers in the NS in the context of a subaccount. Multiple SCCs can be registered in the NS of a single subaccount, as well as, a single SCC can be registered in the NS of multiple subaccounts. Only applications that are running in the subaccount where the SCC is registered can access the on-premise systems exposed by the SCC.
- *Full report* - NS reports to CMP all SCCs that are currently registered in the NS. The report includes all on-premise systems exposed by the SCCs.
- *Delta (incremental) report* - NS reports to CMP only the SCCs, which changed since the previous report.

## Details

This section describes the API schema that a Notification Service must implement to integrate with the NS-Adapter.

### Authorization

In case of a local installation, using the run.sh script, a valid JSON Web Token (JWT) with `tenant.consumerTenant` and `tenant.externalTenant` claims should be provided. The `tenant.consumerTenant` and `tenant.externalTenant` should contain the internal ID and the external ID from the existing `business_tenant_mappings` record in the Compass database.

### Reports
The NS-Adapter gets the following types of reports: 
- `delta` - A regular incremental report on short time intervals, for example, once in 5 minutes.
- `full` - A regular full report on long time periods, for example once per hour.

Each report comprises a list of SCCs. Each SCC contains a list of exposed on-premise systems. When a report comes, the NS-Adapter tries to identify for each SCC the following:

- The systems that are new, to create new entities in CMP for them.
- The systems that are missing, to mark them unreachable in CMP.
- The systems that were updated, to update them in CMP too.

Additionally, when a full report comes, the NS-Adapter identifies if there are SCCs, which are missing in the report, and then, marks all their systems as unreachable.

Each on-premise system stored in CMP is labeled with an `SCC` label with value `{"Subaccount":"{{subaccount of the SCC}}", "LocationID":"{{location-id of the SCC}}", "Host":"{{virtual host of the on-premise system}}"}`. The SCC label is used as a unique identifier for the on-premise systems stored in CMP.

| | |
|---------------------|--------------------------------------|
|**Description**      | Upsert `ExposedSystems` is a bulk `create-or-update` operation on exposed on-premise systems. It takes a list of fully described exposed systems, and then, creates the systems that are new to CMP or updates the metadata for the existing ones.|
|**URL**              | /api/v1/notifications                                  |
|**Query Params**     | reportType=full,delta                                  |
|**HTTP Method**      | PUT                                                    |
|**HTTP Headers**     | Content-Type: application/json                         |
|**HTTP Codes**       | [204 No Content](#204-no-content)                      |
|                     | [200 OK](#200-ok)                                      |
|                     | [400 Bad Request](#400-bad-request)                    |
|                     | [401 Unauthorized](#401-unauthorized)                  |
|                     | [408 Request Timeout](#408-request-timeout)            |
|                     | [500 Internal Server Error](#500-internal-server-error)|
|                     | [502 Bad Gateway](#502-bad-gateway)                    |
|                     | [413 Payload Too Large](#413-payload-too-large)        |
|**Response Formats** | json                                                   |
|**Authentication**   | OAuth 2 via XSUAA                                      |


Example request with `delta` report:
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
          "type": "on-premise-system",
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
          "type": "on-premise-system",
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

Example request with `full` report:
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
          "type": "on-premise-system",
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
          "type": "on-premise-system",
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

This response is returned when all exposed systems in the list are either successfully created or updated. Note that in case of a `full` report, this is the only possible returned response. That is, for a `full` report, the response is returned regardless of whether the `create-or-update` has failed for any SCCs, as these SCCs will be reported again next time.

```
HTTP/1.1 204 No Content
```

#### 200 OK

This response is returned if upsert for one or more on-premise systems has failed. The response is returned only by `delta` reports. It is not returned by a `full` report, as the full report does not return SCCs, for which the update has failed.

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

##### Failed deserialization

This response is returned if the service cannot deserialize the request.

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

This response is returned if the service cannot read the request body from the stream.

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

##### Missing required property

This response is returned in case any of the required properties is not provided, for example, a subaccount.

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
This response is returned in case the mandatory query parameter `reportType` is missing.
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

This response is returned in the event of a failed authentication.

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

This response is returned in the event of a request timeout in the CMP infrastructure.
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

This response is returned in the event of a communication error in the CMP infrastructure.
```
HTTP/1.1 502 Bad Gateway
Content-Type: text/plain

<content>
```
#### 413 Payload Too Large

This response is returned in the event of a request body larger than 5 MB.
```
HTTP/1.1 413 Payload Too Large
Content-Type: text/plain

<content>
```

## Configuration

NS-Adapter binary allows you to override some configuration parameters. To get a list of the configurable parameters of NS-Adapter, see [configuration.go](https://github.com/kyma-incubator/compass/blob/66754e58f6d23f233a6905fa9e4577b6666c98e7/components/director/internal/nsadapter/adapter/configuration.go#L13)

## Local Development

### Prerequisites

NS-Adapter requires access to:
1. A configured PostgreSQL database with imported Director's database schema.
2. A service that exposes endpoint for fetching JWKSs.

### Run
You can start the NS-Adapter via the designated `run.sh` script in Director:
```bash
./run.sh --ns-adapter --jwks-endpoint {your_JWKS_endpoint}

```

This brings up a PostgreSQL database with Director's schema.

The up-to-date properties supported by the `run.sh` script can be found in the Director component local documentation: [Director](../../README.md#local-development).
