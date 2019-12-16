# Mapping

https://kyma-project.io/docs/1.8/components/application-connector/#tutorials-get-the-client-certificate-get-the-csr-information-and-configuration-details-from-kyma

```json
{
    "csrUrl": "{CSR_SIGNING_URL_WITH_TOKEN}",
    "api":{
        "metadataUrl":      "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
        "eventsUrl":        "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
        "infoUrl":          "https://gateway.{CLUSTER_DOMAIN}/v1/applications/management/info",
        "certificatesUrl":  "https://connector-service.{CLUSTER_DOMAIN}/v1/applications/certificates",
    },
    "certificate":{
        "subject":"OU=Test,O=TestOrg,L=Waldorf,ST=Waldorf,C=DE,CN={APP_NAME}",
        "extensions": "",
        "key-algorithm": "rsa2048",
    }
}
```

Compass:

`	requestOneTimeTokenForApplication(id: ID!): OneTimeToken! @hasScopes(path: "graphql.mutation.requestOneTimeTokenForApplication")
`

csrURL = OneTimeToken.connectorURL + OneTimeToken.Token

---
POST /metadata/services

```json
{
  "provider": "example-provider",
  "name": "example-name",
  "description": "This is the long description of your service",
  "shortDescription": "Short description",
  "labels": {
    "example": "true"
  },
  "api": {
    "targetUrl": "https://httpbin.org/",
    "spec": {},
    "requestParameters": {
      "queryParameters": {
        "param": ["bar"]
      },
      "headers": {
        "custom-header": ["foo"]
      }
    },
    "credentials": {
      "basic": {
        "username": "{USERNAME}",
        "password": "{PASSWORD}"
      }
    }
  },
  "events": {
    "spec": {
      "asyncapi": "1.0.0",
      "info": {
        "title": "PetStore Events",
        "version": "1.0.0",
        "description": "Description of all the events"
      },
      "baseTopic": "stage.com.some.company.system",
      "topics": {
        "petCreated.v1": {
          "subscribe": {
            "summary": "Event containing information about new pet added to the Pet Store.",
            "payload": {
              "type": "object",
              "properties": {
                "pet": {
                  "type": "object",
                  "required": [
                    "id",
                    "name"
                  ],
                  "example": {
                    "id": "4caad296-e0c5-491e-98ac-0ed118f9474e",
                    "category": "mammal",
                    "name": "doggie"
                  },
                  "properties": {
                    "id": {
                      "title": "Id",
                      "description": "Resource identifier",
                      "type": "string"
                    },
                    "name": {
                      "title": "Name",
                      "description": "Pet name",
                      "type": "string"
                    },
                    "category": {
                      "title": "Category",
                      "description": "Animal category",
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "documentation": {
    "displayName": "Documentation",
    "description": "Description",
    "type": "some type",
    "tags": ["tag1", "tag2"],
    "docs": [
        {
        "title": "Documentation title...",
        "type": "type",
        "source": "source"
        }
    ]
  }
}
```

Compass:
`addAPIDefinition`
`addEventDefinition`
TODO: we don't have such thing as a service


TODO:
> To register an API with a specification URL, replace api.spec with api.specificationUrl.

SpecificationURL - FetchRequests are not implemented


> Register an OData API
> If the api.spec or api.specificationUrl parameters are not specified and the api.type parameter is set to OData, the Application Registry will try to fetch the specification from the target URL with the $metadata path

Special handling of OData API can be done in the Adapter, to create FethRequest

GET /metadata/services
GET /metadata/services{serviceID}

Compass:
getApplication (how do we know which one?), find proper API Definition and Event Definition.
But we cannot mark anyway that this API Definition and Event Definition
belongs to the same Service. API Packages has to be implemented.