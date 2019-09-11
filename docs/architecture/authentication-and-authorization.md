# Authentication and Authorization

## Introduction
Currently, communication between the Compass and both runtimes and applications are not secured. We need to provide some security possibilities.
We want to secure the Compass using ORY's Hydra and Oathkeeper. There would be two ways of authentication:
 - Oauth 2.0 
 - SSL/TLS Certificates.

## Proposed solution
As mentioned, there would be two possibilities for securing the connection. To achieve that, first, we need to integrate Hydra and Oathkeeper into the Compass. Then we will introduce a new component - Mapping Service.
Mapping Service would be a part of Management Plane Services, responsible for extracting from database a `tenant` based on `client_id`.
We would also use a Hydrator and id_token service which are ORY's mutators.

Having that setup, the flow would be as follows: 

1. Runtime/Application calls the Oathkeeper.
2. Oathkeeper calls Hydrator with `client_id` attached to request.
3. Hydrator calls Mapping Service and gets `tenant` in response.
4. Hydrator calls id_token service with `tenant` attached to the request
5. The id_token service constructs a JWT token with proper scopes and `tenant` injected inside it.
6. The token issued by id_token service can be used to authenticate to the Gateway.

![Auth](./assets/compass-auth.svg)

### Architecture

#### Mapping service
It is a service responsible for mapping `client_id` (Oauth2) to `tenant`. It would use the same database which uses the Director component. 

#### Managing tenants
We will need to have a table in the database which will consist of `tenant` and `client_id`

#### Oauth2
Request flow:
1. Oathkeeper calls Hydra for introspection of token
2. If the token was valid, Oathkeeper sends request to Hydrator 
3. Hydrator calls Mapping Service to get `tenant` from a `client_id`
4. Hydrator calls ID_Token mutator to create a JWT token with injected `tenant` field
5. ID_Token mutator calls the Compass Gateway
   
![Auth](./assets/oauth2-diagram.svg)

#### Certificates
TODO

#### Communication between Runtime and Application
TODO

## Summary

### Oauth2
Things to be done: 
- Integrate Hydra and Oathkeeper into the Compass [(POC)](https://github.com/kyma-incubator/compass/issues/290)
- Managing tenants (storing `tenant` with `client_id` in the database)
- Mapping service which will be responsible for mapping `client_id` to the `tenant`, it will use the same database as the Director