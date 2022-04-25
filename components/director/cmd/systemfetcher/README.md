# System Fetcher

## Overview

This application fetches customer systems from an external service. Then, the fetched systems are synchronized with the Compass database.

The purpose of the system fetcher is to populate customer systems automatically, instead of creating these systems manually using the UI, or using the Director GraphQL API.

## Details

This section describes the API schema that must be implemented on the server side to integrate with the System Fetcher.

### Authorization

The External System Registry API uses either OAuth 2.0 client credentials authorization flow or mTLS, and comprises preconfigured trust for the externally issued client certificate of Compass.

### Systems Endpoint

The endpoint must return a specific payload and must accept the following type of query parameters:
- **global.systemFetcher.systemsAPIFilterCriteria** - Specifies the filtering criteria for the systems. For example, the system type (mapped to an application template in Compass).
- **global.systemFetcher.systemsAPIFilterTenantCriteriaPattern** - Specifies the filtering criteria for the systems, based on a tenant from Compass.

#### Response

The response of the system API returns the following response:

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
Then, using the input, the System Fetcher can create a system by template.

## Configuration

The System Fetcher binary allows you to override some configuration parameters. To get a list of the configurable parameters, see [main.go](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/systemfetcher/main.go#L48).

## Local Development
### Prerequisites
The System Fetcher requires access to:
1. Configured PostgreSQL database with the imported Director's database schema.
1. API that can be called to fetch systems. For details about implementing the System Registry API that the System Fetcher can consume, see the [Systems Endpoint](#systems-endpoint) section in this document. 

### Run
Since the System Fetcher is usually a short-lived process, it is useful to start and debug it directly from your IDE.
Make sure that you provide all required configuration properties as environment variables.
