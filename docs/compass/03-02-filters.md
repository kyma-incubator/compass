# Envoy Filters

There are a couple of Envoy filters used for various purposes - header manipulation, security, validation.

### Correlation Headers Rewrite Filter
Correlation IDs are used for uniquely identifying requests across microservices. The origin microservice creates the correlation ID, and it is attached as a header to all requests that might follow.
Istio also provides a tracing mechanism, which is described in details [here](https://istio.io/latest/about/faq/distributed-tracing/#how-to-support-tracing). The most important thing is that it uses `x-request-id` to do so, hence `x-request-id` is the correlation header we will use for requests across our components.

In case an external service has attached a correlation ID header, we will get its value, and set it to `x-request-id` - that's what the `compass-correlation-headers-rewrite` Envoy filter does.
It is required because there is no strict convention regarding the correlation ID header name. The default ones we're looking for are listed [here](https://github.com/kyma-incubator/compass/blob/d429d01b34eb1a7512c9613dff2bc27d6c814857/chart/compass/values.yaml#L304).

### Disable Istio Retries Filter
This filter sets the `x-envoy-max-retries` header to 0, which will disable the default retry behavior, because it is not flexible enough with the currently supported Istio version. 

### ORY Oathkeeper Timeout Filter
This filter was introduced because there are no timeouts on ORY Oathkeeper side, but we want to cancel requests that take more time than usual.
This filter is rather a workaround for missing Oathkeeper functionality. More details can be found in [the PR](https://github.com/kyma-incubator/compass/pull/1886) that introduced the filter.

### Limit Request Payload Size Filter
As the name suggest, this filter is used for limiting the request payload size for POST requests. It was introduced as a security measure, as requests with huge payload may cause denial of service.

### Rewrite Token Filter
This filter takes care for requests coming to the Pairing Adapter (the REST Adapter for Director). In those cases, if a One-Time Token is used for pairing, it is passed as a query parameter in the request. This filter takes care of extracting it from the query request, and adding it as a `Token` header.

### Rewrite Client Certificate Filter
This filter is responsible for populating the certificate data header (default key is `Certificate-Data`) with the contents of the Envoy proxy header [`X-Forwarded-Client-Cert`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-client-cert) which indicates certificate information of part or all of the clients or proxies that a request has flowed through, on its way from the client to the server.

### Preserve External Host
This filter takes care of persisting the `Host` header of incoming requests into a separate `x-external-host` header.

### Rate Limiting Filters
The last filters are related to rate limiting. They are documented in details [here](./10-01-rate-limiting.md)
