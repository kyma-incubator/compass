# Application APIs and Events Discovery

Similarly to the applications (systems) discovery, there are a couple of ways, in which Compass finds the APIs and events provided by an application:
- **Manually** - Register APIs and events via the Director's GraphQL API.
- **Automatically** - Discovered automatically if they have an appropriate webhook for that.

The ORD Aggregator component takes care of that discovery. It is modeled as a Kubernetes CronJob. It periodically synchronizes the available APIs, events, and specifications of a given application with the actually available ones. Presumably, the actual state is dynamic and APIs are made available (or removed) automatically, and this does not require manual updates on Compass side.

## Resource Discovery Aggregator
The aggregator processes all applications at one CronJob run. If the application has a webhook for resource discovery, the aggregator synchronizes the application's vendors, bundles, packages, their APIs, and event definitions along with their specifications. That is, the required resources are either created or removed from the Compass database.

### ORD Webhooks
The webhook that is called by the Aggregator is a standard Compass webhook. It returns a list of documents called Resource Discovery (RD) Documents.
Each RD Document contains a list of bundles, APIs, and event definitions. Then, to fetch the specifications of the APIs and events, the Aggregator uses different access strategies:
- Open - Indicates that the document is not secured.
- CMP mTLS - Indicates that the document can be accessed with Compass's externally issued certificate.

## Resource Discovery Service

The ORD Service component is developed in a separate GitHub repository at [ORD Service](https://github.com/kyma-incubator/ord-service).

It exposes an OData API that fetches read-only information about the applications that are available in Compass. The inoframtion includes application vendors, bundles, packages, APIs, and event definitions.
