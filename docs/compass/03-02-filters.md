# Envoy Filters

There are several Envoy filters used for various purposes, such as, header manipulation, security, or validation.

### Correlation Headers Rewrite Filter
Correlation IDs are used for uniquely identifying requests across microservices. The origin microservice creates the correlation ID, and it is attached as a header to all requests that might follow.
Istio also provides a tracing mechanism, which is described in details at [Distributed Tracing](https://istio.io/latest/about/faq/distributed-tracing/#how-to-support-tracing). The most important thing is that it uses `x-request-id` to do so, hence `x-request-id` is the correlation header that is used for requests across the Compass components.

When an external service has attached a correlation ID header, its value is get and set to `x-request-id`. This is what the `compass-correlation-headers-rewrite` Envoy filter does.
It is required because there is no strict convention regarding the correlation ID header name. The default names that are processed are listed at [values.yaml](https://github.com/kyma-incubator/compass/blob/d429d01b34eb1a7512c9613dff2bc27d6c814857/chart/compass/values.yaml#L304).

### Disable Istio Retries Filter
This filter sets the `x-envoy-max-retries` header to 0. It disables the default retry behavior, because it is not flexible enough with the currently supported Istio version. 

### ORY Oathkeeper Timeout Filter
This filter was introduced because there are no timeouts on ORY Oathkeeper side. Yet, requests that take more time than usual must be canceled.
This filter is rather a workaround for missing Oathkeeper functionality. For more information, see the pull request that introduced the filter: [Pull Request 1886](https://github.com/kyma-incubator/compass/pull/1886).

### Limit Request Payload Size Filter
This filter is used for limiting the request payload size for POST requests. It was introduced as a security measure as requests with huge payload can cause denial of service.

### Rewrite Token Filter
This filter takes care for requests coming to the Pairing Adapter (the REST Adapter for Director). In those cases, if an One-Time Token is used for pairing, it is passed as a query parameter in the request. This filter takes care of extracting it from the query request and adding it as a `Token` header.

### Rewrite Client Certificate Filter
This filter populates the certificate data header (default key is `Certificate-Data`) with the contents of the Envoy proxy header [`X-Forwarded-Client-Cert`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-client-cert). It indicates the certificate information (of a part or all) of the clients or proxies that a request has flowed through on its way from the client to the server.

### Preserve External Host
This filter persists the `Host` header of incoming requests into a separate `x-external-host` header.

### Rate Limiting Filters
The last filters perform rate limiting. They are documented in details at [Rate Limiting](./10-01-rate-limiting.md)
