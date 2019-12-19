# Provide REST Adapter for Compass Graphql API
Compass GraphQL API is not compatible with Application Connector REST API.
To avoid rewriting integration layer for  Application Connector clients, we can provide adapter that translates
Application Connector API to Compass API.

1. Authorization and Authentication should be the same as for thr Application Connector API.
2. Adapter then is registered as a Integration System, so it should have access to all Graphql mutations and queries.
3. Consider following example:
` POST /v1/applications/certificates/revocations` which revokes current certificate.

- Client calls Adapter with certificate to revoke.
- Istio validates if call to Adapter is valid.
- Istio removes information about certificate, adds header with hash certificate
- Adapter calls with their own certificate, but forward header with hash certificate

# Application Registry API

## Adding API
```
curl -k --cert ./app1_client.crt   --key ./generated.key -X POST https://gateway.34.77.12.120.xip.io/app1/v1/metadata/services --data-binary "@serviceRegistrationManyAPIs.json"
```
```json
{
  "provider":"SAP",
  "name":"api1",
  "description": "api1 desc",
  "shortDescription": "api1 short desc",
  "identifier": "api1Identifier",
  "labels":{
  	"label1":"value1",
  	"label2": "value2"
  },
  "api":{
  	"targetUrl":"http://some-app1.pl",
  	"requestParameters": {
  	  "headers": {
  	      "header1": "value1"
  	  }
  	}
  }

}
```
Compass equivalent: `addApiDefinition`


- in old API, there is a Service which seems to be equivalent to API Package, which is missing in Compass
- `identifier` has no mapping in Compass API
- you cannot label service class 
- specificationURL, specificationCredentials, specificationRequestParameters can be mapped to FetchRequest, but it is not yet implemented
- from Documentation:
> If the api.spec or api.specificationUrl parameters are not specified and the api.type parameter is set to OData, the Application Registry will try to fetch the specification from the target URL with the $metadata path.

Special handling for OData can be implemented in the Adapter (assuming that FetchRequest is implemented)

- with adding events it should be no problem, because Application Registry API is very simple and requires only Event Spec.
- documentation: in Compass, documentation can be added only on the Application level

## Get all registered services GET /v1/metadata/services
Compass equivalent: `application({{ID}}) {apiDefinitions {...} }`

## Get a service by service ID
We don't have serviceID in Compasss API, but when we got list of IDs we can execute:
`application({{ID}}) { apiDefinition({{API_ID}}) {...} }`
The same comment applies for update and delete described below.

## Updates a service by service ID
Compass equivalent: `updateAPIDefinition(id: {{API_ID}}, in: {...})`

## Deletes a service by service ID
Compass equivalent: `deleteAPIDefinition(id: {{API_ID}})`

# Application Connector

## POST /v1​/applications​/certificates​/revocations Marks certificate as revoked.
Compass Connector equivalent: `revokeCertificate`
## GET /v1​/applications​/signingRequests​/info Allows for fetching information for CSR.
Compass Connector equivalent: ``
## GET /v1​/applications​/management​/info Returns information on available services.
Compass Connector equivalent: ``
## POST /v1​/applications​/certificates Signs CSR.
Compass Connector equivalent: ``
## POST /v1​/applications​/certificates​/renewals Renews certificate using CSR.
Compass Connector equivalent: ``

## runtimes external API do we need this at all?
