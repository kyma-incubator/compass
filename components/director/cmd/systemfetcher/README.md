# System Fetcher

## Overview

This application fetches customer systems from an external service.
Afterwards, the fetched systems are synchronized with the Compass database.

The idea is to automatically populate customer systems instead of having them created manually through our UI or through our Director GraphQL API.

## Details

This section describes the API schema that a server must implement to integrate with the System Fetcher.

### Authorization

The External System Registry API should use the OAuth 2.0 client credentials authorization flow, or mTLS and trust the Compass client certificate.

### Systems Endpoint

The endpoint must return a specific payload and accept the following type of query parameters:
- **global.systemFetcher.systemsAPIFilterCriteria** - specifies the criteria on which systems can be filtered, for example their type (mapped to a application template in Compass)
- **global.systemFetcher.systemsAPIFilterTenantCriteriaPattern** - specifies the criteria on which systems can be filtered based on a tenant from Compass

#### Response

The response of the systems API should return the following response:

```json
[
    {
        "systemNumber": "<unique-id>",
	    "displayName": "<name>",
	    "productDescription": "<description>",
	    "baseUrl": "<baseURL>",
	    "infrastructureProvider": "<provider>",
	    "additionalUrls": "<additional-urls>",
	    "additionalAttributes": "<additional-attributes>"
    }
]
```
Then the System Fetcher is able to create an application from a template with that input.

## Configuration

The System Fetcher binary allows overriding of some configuration parameters. Up-to-date list of the configurable parameters can be found [here](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/systemfetcher/main.go#L48).

## Local Development
### Prerequisites
The System Fetcher requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
1. API that can be called to fetch applications. For details about implementing the System Registry API that the System Fetcher can consume, see the [Systems Endpoint](#systems-endpoint) section of this document. 

### Run
Since the ORD Aggregator is usually a short-lived process, it is useful to start and debug it directly from your IDE.
Make sure to provide all required configuration properties as environment variables.
