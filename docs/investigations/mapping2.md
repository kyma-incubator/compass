# Mapping

## Connector API Migration

| Connector API                                 | Compass Adapter                         |
|-----------------------------------------------|-----------------------------------------|
| POST /v1/applications/tokens/                 | Internal API, no action required        |
| POST /v1/runtimes/tokens                      | API for runtimes, no action required    |
| POST /v1​/runtimes​/certificates​/revocations    | as above |
| GET /v1/runtimes/signingRequests/info         | as above
| GET /v1/runtimes/management/info              | as above
| POST /v1/runtimes/certificates                | as above
| POST /v1​/runtimes​/certificates​/renewals       | as above
| POST /v1/applications/certificates/revocations | 
| GET /v1/applications/signingRequests/info     |   
| GET /v1/applications/management/info          |
| POST /v1/applications/certificates            |
| POST /v1/runtimes/certificates/renewals       |


### POST /v1/applications/certificates/revocations
2. Graphql
```graphql
query {
  application(id:{APPLICATION_ID}}) {
    auths {
      id
    }
  }
}

```

Then, for every `SystemAuth`
```graphql
mutation {
  deleteSystemAuthForApplication(authID: {{AUTH_ID}})
}
```

### GET /v1/applications/signingRequests/info 
1. REST:
```json

{
  "clientIdentity": {
    "application": "{APP_NAME}",
    "group": "",
    "tenant": ""
  },
  "urls": {
    "metadataUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
    "eventsUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
    "renewCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/renewals",
    "revokeCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations"
  },
  "certificate": {
    "subject": "OU=Test,O=Test,L=Blacksburg,ST=Virginia,C=US,CN={APP_NAME}",
    "extensions": "string",
    "key-algorithm": "rsa2048"
  }
}
```
 
2. `connector/graphql`:
```graphql
query {
    result: configuration {
        token {
            token
        }
        certificateSigningRequestInfo {
            subject
            keyAlgorithm
        }
        managementPlaneInfo {
            directorURL
            certificateSecuredConnectorURL
        }
    }
}
```

TODO: urls to Compass Adapter should be provided, and secured by client certificate.
`clientIdentity` is a `viewer` in Compass.

### GET /v1/applications/management/info
Seems to be simlar to `signingRequest/info`

### POST /v1/applications/certificates

Graphql: `signCertificateSigningRequest`

### POST /v1/runtimes/certificates/renewals 

Missing counter-part in Compass.


## Application Registry API Migration

| Application Registry                          | Compass Adapter                       |
|-----------------------------------------------|---------------------------------------|
| POST /v1/metadata/services                    |
| GET /v1/metadata/services                     |
| GET /v1/metadata/services/{serviceId}         |
| PUT /v1/metadata/services/{serviceId}         |
| DELETE /v1/metadata/services/{serviceId}      |

### POST /v1/metadata/services 
1. REST
```json
{
  "provider": "string",
  "name": "string",
  "description": "string",
  "shortDescription": "string",
  "identifier": "string",
  "labels": {
    "additionalProp1": "string",
    "additionalProp2": "string",
    "additionalProp3": "string"
  },
  "api": {
    "targetUrl": "string",
    "credentials": {
      "oauth": {
        "url": "string",
        "clientId": "string",
        "clientSecret": "string"
      },
      "basic": {
        "username": "string",
        "password": "string"
      },
      "certificateGen": {
        "commonName": "string",
        "certificate": "string"
      }
    },
    "spec": {},
    "SpecificationUrl": "string",
    "ApiType": "string",
    "requestParameters": {
      "headers": {
        "additionalProp1": [
          "string"
        ],
        "additionalProp2": [
          "string"
        ],
        "additionalProp3": [
          "string"
        ]
      },
      "queryParameters": {
        "additionalProp1": [
          "string"
        ],
        "additionalProp2": [
          "string"
        ],
        "additionalProp3": [
          "string"
        ]
      }
    },
    "specificationCredentials": {
      "oauth": {
        "url": "string",
        "clientId": "string",
        "clientSecret": "string"
      },
      "basic": {
        "username": "string",
        "password": "string"
      }
    },
    "specificationRequestParameters": {
      "headers": {
        "additionalProp1": [
          "string"
        ],
        "additionalProp2": [
          "string"
        ],
        "additionalProp3": [
          "string"
        ]
      },
      "queryParameters": {
        "additionalProp1": [
          "string"
        ],
        "additionalProp2": [
          "string"
        ],
        "additionalProp3": [
          "string"
        ]
      }
    }
  },
  "events": {
    "spec": {}
  },
  "documentation": {
    "displayName": "string",
    "description": "string",
    "type": "string",
    "tags": [
      "string"
    ],
    "docs": [
      {
        "title": "string",
        "type": "string",
        "source": "string"
      }
    ]
  }
}
```

2. GraphQL

We don't have a such object as a servce in Compass GraphQL API, probably we have to wait 
until we implement packageAPI.
For now, we have following mutations:
```graphql
addAPIDefinition
addEventDefinition
```
TODO: how to provide documentation for PackageAPI?

## Events API Migration

| Events API                                    | Compass Adapter                       |
|-----------------------------------------------|---------------------------------------|
| GET /{application}/v1/events/subscribed       |   
| POST /{application}/v1/events                 |