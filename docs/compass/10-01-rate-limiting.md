# CMP Rate Limiting

Rate-limiting is implemented in CMP using the [Istio's local rate-limiting](https://istio.io/latest/docs/tasks/policy-enforcement/rate-limit/#local-rate-limit) functionality which provides enabling rate-limiting at instance level. In Kubernetes terms, that can be thought of as rate-limiting per Deployment. Enablement is done by applying an [EnvoyFilter](https://istio.io/latest/docs/reference/config/networking/envoy-filter/) which is responsible for maintaining **token-buckets**. A token-bucket is a simple data structure which contains a fixed amount of tokens. For each HTTP request that arrives a token is removed from the token-bucket. In case the token-bucket gets emptied, the next incoming request gets rate-limited causing an early termination of the request handling and returning an HTTP status-code of 429 (Too Many Requests). The token bucket gets refilled on regular intervals with a fixed amount of tokens. The initial token count, refill rate and the amount of tokens per refill can be configured in the EnvoyFilter's resource description.

## EnvoyFilters in CMP 

We have created 2 rate-limiting *EnvoyFilter* resources:

- *EnvoyFilter* for **Compass Gateway**
- *EnvoyFilter* for **Compass ORD-Service**

Since most of our auth flows pass through the Gateway component (the upstream for the majority of our [ORY rules](../../chart/compass/charts/gateway/templates/oathkeeper-rules.yaml) is the gateway service) it is an appropriate service to apply rate-limiting which captures most of the incoming requests to CMP. Several flows end up in the ord-service and that is why the second *EnvoyFilter* is applied there. The rest of the flows are either for internal calls (no rate-limiting needed) or for the system-broker component for which CMP does not apply rate-limiting.
 
### Gateway EnvoyFilter

The [Gateway EnvoyFilter](../../chart/compass/charts/gateway/templates/rate-limit-filter.yaml) leverages [descriptors](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/local_rate_limit_filter#using-rate-limit-descriptors-for-local-rate-limiting) and thus describes 3 token buckets. A *descriptor* is a key-value pair which provides information about the current request. *Descriptors* are created by *actions*. 

The gateway EnvoyFilter has 1 action which creates a descriptor with a key `consumer`. The value of that descriptor for each HTTP request is retrieved from an HTTP header which in the default configuration is `X-Flow-Identity`. This header is added in the **compass-director-certs** and **compass-connector-certs** auth flows (described by the [ORY rules](../../chart/compass/charts/gateway/templates/oathkeeper-rules.yaml) with the same names) by an ORY Header Hydrator which inspects the ORY session to get the value for it. The value of the header is the type of the consumer (*Application*, *Runtime*  or *Integration System*). Thus, the `consumer` descriptor's value can be one of those 3 possible consumer types. 

In case the value of the descriptor is *Runtime* or *Application* the request will get matched by the EnvoyFilter and a separate token-bucket will be used to determine whether to rate-limit the request or not. Otherwise, the *general* token-bucket will be used. Notice that all other auth flows (except from **compass-director-certs** and **compass-connector-certs**) match [ORY rules](../../chart/compass/charts/gateway/templates/oathkeeper-rules.yaml) which don't add the `X-Flow-Identity` header, respectively the `consumer` descriptor is not created for them meaning that they also end up in the *general* token-bucket. The following diagram depicts the flow:

![](./assets/gateway-rate-limiting.svg)

The header that is used to fill the value with the descriptor is configurable form the values of the compass Helm chart.

### ORD-Service EnvoyFilter

The flows which end up in the **Compass ORD-Service** without passing through the **Compass Gateway** are processed by the [ORD-Service EnvoyFilter](../../chart/compass/charts/ord-service/templates/rate-limit-filter.yaml). This EnvoyFilter uses a simpler version of the token-bucket method described for the Gateway. In this scenario there is **only one** token-bucket and the rate-limiting for all incoming HTTP requests is determined by it. 