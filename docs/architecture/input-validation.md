# Input Validation

## Overview

This document contains validation rules for all input types.

## Validation rules explanation

- `printable` - Printable unicode characters.
- `printableWithWhitespace` - Printable unicode characters and whitespace characters.
- `name` - Up to 36 characters long. The characters allowed in names are: digits (`0`-`9`), lower case letters (`a`-`z`),`-`, and `.`. Based on Kubernetes resource name format.
- `required` - Cannot be nil or empty.
- `not_empty` - Cannot be empty (can be nil if pointer).
- `url` - Valid URL.
- `max` - Maximal allowed length.
- `oneof` - Value has to be one of specified values.
- `uuid` - Valid UUID.


## Proposed validation rules for Compass input types

### APIDefinitionInput

Field | Rules | Comment
--- | --- | ---
name: String! | `required`,`name`|  
description: String | `not_empty`,`max=128`,`printableWithWhitespace` |  
targetURL: String! | `required`,`url`,`printable`,`max=256` |  varchar(256) in db
group: String |  `not_empty`,`printable`,`max=36` |  varchar(256) in db
spec: APISpecInput |   | 
version: VersionInput |   |  
defaultAuth: AuthInput |   |  

### APISpecInput

- Struct validator ensuring that `type` and `format` work together (ODATA works with XML and JSON, OPEN_API works with YAML and JSON)

Field | Rules | Comment
--- | --- | ---
data: CLOB (string) | `not_empty`,`printableWithWhitespace` |  
type: APISpecType! | `required`,`oneof=ODATA OPEN_API`,`printable` |  
format: SpecFormat! | `required`,`oneof=YAML JSON XML`,`printable` |  
fetchRequest: FetchRequestInput |   |  

### EventAPIDefinitionInput

Field | Rules | Comment
--- | --- | ---
name: String! | `required`,`name` | varchar(256) in db  
description: String | `not_empty`,`printableWithWhitespace`,`max=128` |  
spec: EventAPISpecInput! | `required` | 
group: String | `not_empty`,`printable`,`max=36`  | varchar(256) in db  
version: VersionInput |   |  

### EventAPISpecInput

- ~~Struct validator ensuring that `type` and `format` work together (ASYNC_API works with YAML and JSON)~~ not needed yet because we've only one event api spec type

Field | Rules | Comment
--- | --- | ---
data: CLOB (string) | `not_empty`,`printableWithWhitespace` |  
eventSpecType: EventAPISpecType! | `required`,`oneof=ASYNC_API`,`printable` |  
format: SpecFormat! | `required`,`oneof=YAML JSON`,`printable` |  
fetchRequest: FetchRequestInput |   |  

### VersionInput

Field | Rules | Comment
--- | --- | ---
value: String! | `required`,`printable`,`max=256` | varchar(256) in db
deprecated: Boolean = false | `required` | required because has default value (?)
deprecatedSince: String | `not_empty`,`printable`,`max=256` | varchar(256) in db
forRemoval: Boolean = false | `required` | required because has default value (?)

### ApplicationCreateInput

Field | Rules | Comment
--- | --- | ---
name: String! | `required`,`name` | max 36 characters so frogs are able to append suffix
description: String | `not_empty`,`printableWithWhitespace`,`max=128` |  
labels: Labels (map[string]interface{}) |  |  
webhooks: [WebhookInput!] | `[required]` |  
healthCheckURL: String | `not_empty`,`url`,`printable`,`max=256` | varchar(256) in db  
apis: [APIDefinitionInput!] | `[required]` |  
eventAPIs: [EventAPIDefinitionInput!] | `[required]` |  
documents: [DocumentInput!] | `[required]` |  

### ApplicationUpdateInput

Field | Rules | Comment
--- | --- | ---
name: String! | `required`,`name` | max 36 characters so frogs are able to append suffix
description: String | `not_empty`,`printableWithWhitespace`,`max=128` |  
healthCheckURL: String | `not_empty`,`url`,`printable`,`max=256` | varchar(256) in db  

### RuntimeInput

Field | Rules | Comment
--- | --- | ---
name: String! | `required`,`name` | varchar(256) in db
description: String | `not_empty`,`printableWithWhitespace`,`max=128` |
labels: Labels (map[string]interface{}) | |

### IntegrationSystemInput

Field | Rules | Comment
--- | --- | ---
name: String! | `required`,`name` | varchar(256) in db  
description: String | `not_empty`,`printableWithWhitespace`,`max=128` |  

### DocumentInput

Field | Rules | Comment
--- | --- | ---
title: String! | `required`,`printable`,`max=128` |  varchar(256) in db
displayName: String! | `required`,`printable`,`max=128` |  varchar(256) in db
description: String! | `required`,`printableWithWhitespace`,`max=128` |  
format: DocumentFormat! | `required`,`printable`,`oneof=MARKDOWN` |  
kind: String | `not_empty`,`printable`,`max=256`  |  varchar(256) in db
data: CLOB (string) | `not_empty`,`printableWithWhitespace` |  
fetchRequest: FetchRequestInput |  |  

### WebhookInput

Field | Rules | Comment
--- | --- | ---
type: ApplicationWebhookType! | `required`,`printable`,`oneof=CONFIGURATION_CHANGED` |
url: String! | `required`,`url`,`printable`,`max=256` | varchar(256) in db
auth: AuthInput | |

### LabelDefinitionInput

Field | Rules | Comment
--- | --- | ---
key: String! | `required`,`printable`,`max=256` | varchar(256) in db  
schema: JSONSchema (string) | `not_empty`,`printableWithWhitespace` |  

### LabelInput

Field | Rules | Comment
--- | --- | ---
key: String! | `required`,`printable`,`max=256` | varchar(256) in db  
value: Any! (interface{}) | `required` | 

### FetchRequestInput

Field | Rules | Comment
--- | --- | ---
url: String! | `required`,`url`,`printable`,`max=256` | varchar(256) in db  
auth: AuthInput |  |  
mode: FetchMode = SINGLE | `required`,`oneof=SINGLE PACKAGE INDEX`,`printable` | required because has default value (?)
filter: String | `not_empty`,`printable`,`max=256`  | varchar(256) in db  

### AuthInput

Field | Rules | Comment
--- | --- | ---
credential: CredentialDataInput! | `required` |  
additionalHeaders: HttpHeaders (map[string][]string) |   |  
additionalQueryParams: QueryParams (map[string][]string) |   |  
requestAuth: CredentialRequestAuthInput |   | 

### CredentialDataInput

- Struct validator ensuring that exactly one field is not nil

Field | Rules | Comment
--- | --- | ---
basic: BasicCredentialDataInput |   |  
oauth: OAuthCredentialDataInput |   |  

### BasicCredentialDataInput

Field | Rules | Comment
--- | --- | ---
username: String! | `required`,`printable` |  
password: String! | `required`,`printable` |  

### OAuthCredentialDataInput

Field | Rules | Comment
--- | --- | ---
clientId: ID! | `required`,`printable`,`uuid` |
clientSecret: String! | `required`,`printable` |
url: String! | `required`,`printable`,`url` |

### CredentialRequestAuthInput

- Struct validator ensuring that exactly one field is not nil

Field | Rules | Comment
--- | --- | ---
csrf: CSRFTokenCredentialRequestAuthInput |   |  

### CSRFTokenCredentialRequestAuthInput

Field | Rules | Comment
--- | --- | ---
tokenEndpointURL: String! | `required`,`printable`,`url` |  
credential: CredentialDataInput! | `required` | 
additionalHeaders: HttpHeaders (map[string][]string) |   |  
additionalQueryParams: QueryParams (map[string][]string) |   |  
