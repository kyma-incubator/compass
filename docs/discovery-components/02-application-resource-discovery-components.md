# Discovering Application APIs and Events

Just like with applications (also called _systems_), there are two ways for Compass to find out the APIs and events provided by an application - they can be registered **manually** via the Director GraphQL API, or they can be **automatically** discovered, if they have the appropriate webhook for that.

The ORD Aggregator component takes care of that discovery. It is modeled as a Kubernetes CronJob. It periodically synchronizes the available APIs, events, and specifications of a given application, with the actually available ones - it is assumed that the actual state is dynamic, and for example, APIs can be made available, or removed, and that should not require manual updates on Compass side.

## Resource Discovery Aggregator
The aggregator processes all applications on one CronJob run. If the application has a webhook for resource discovery, the aggregator synchronizes the application's vendors, bundles, packages, their API and event definitions along with their specifications.
By synchronize it's meant that the required resources are created or removed from the Compass database.

### ORD Webhooks
The webhook called from the Aggregator is a standard Compass webhook. It returns a list of the so-called _Resource Discovery Documents_.
Each RD Document contains a list of bundles, API and event definitions. Then, in order to fetch the specifications of the APIs and events, the Aggregator can use different access strategies:
* Open - indicates that the document is not secured
* CMP mTLS - indicates that the document can be accessed with Compass' Externally issued certificate

## Resource Discovery Service

The ORD Service component is developed in a separate GitHub [repository](https://github.com/kyma-incubator/ord-service).

It exposes is a read-only OData API for the applications' information available in Compass - application vendors, bundles, packages, API and event definitions.
