# Pairing Application with Compass

## Introduction

Application is one of the three basic objects in Compass. It is an abstraction over the actual Application instance. To fully benefit from the Compass features, the Application instance must be paired with its abstraction. For that reason, the Compass offers the API.

## API specification

The `AuthTypeInput` GraphQL input parameter describes the supported authentication type, for example OAuth2 or ClientCertificate.

```graphql
input AuthTypeInput {
    type: string!
}
```

The `PairingAuthResult` GraphQL type wraps the result returned from the `generateAuthForApplication` mutation. It wraps the authentication details, such as `client_id`, `client_secret`, or a one-time token.

```graphql
type PairingAuthResult {
    params: JSON!
}
```

The `AuthType` GraphQL type wraps the requested authentication type.

```graphql
type AuthType {
    type: string!
}
```

The `getAuthTypesForApplication(appID: ID!): [AuthType!]!` query returns the possible values for the `type` parameter for auth generation mutation.

```graphql
type Query {
    getAuthTypesForApplication(appID: ID!): [AuthType!]!
}
```

The `generateAuthForApplication(appID: ID!, type: AuthTypeInput!, inputParams: JSON!): PairingAuthResult!` mutation is responsible for providing the authentication details for the Application to perform the pairing.

```graphql
type Mutation {
    generateAuthForApplication(appID: ID!, type: AuthTypeInput!, inputParams: JSON!): PairingAuthResult!
}
```

## Establishing a trusted relation for the Application

Compass offers two ways of establishing a trusted relation:
1. [Pairing without use of Integration System](./pairing-without-using-integration-system.md)
1. [Pairing with use of Integration System](./pairing-with-use-integration-system.md)
