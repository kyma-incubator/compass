# Input Validation

## Overview

This document contains validation rules for all input types.

## Validation rules explanation

- `name` - Up to 36 characters long. Cannot start with digit. The characters allowed in names are: digits (`0`-`9`), lower case letters (`a`-`z`),`-`, and `.`. Based on Kubernetes resource name format.
- `required` - Cannot be nil or empty.
- `url` - Valid URL.
- `max` - Maximal allowed length.
- `oneof` - Value has to be one of specified values.
- `[$VALIDATION_RULE]` - Array that can be nil or empty but every array element has to fulfill specified `$VALIDATION_RULE`.


## Proposed validation rules for Compass input types

### APIDefinitionInput

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` |  
description: String | false |`max=128` |  
targetURL: String! | true | `url`, `max=256` |  varchar(256) in db
group: String | false | `max=36` |  varchar(256) in db
spec: APISpecInput | false | | 
version: VersionInput | false | |  
defaultAuth: AuthInput | false | |  

### APISpecInput

- Struct validator ensures that `type` and `format` work together (ODATA works with XML and JSON, OPEN_API works with YAML and JSON)
- Struct validator ensures that only one of `data` and `fetchRequest` is present

Field | Required | Rules | Comment
--- | --- | --- | ---
data: CLOB (string) | false | |  
type: APISpecType! | true | `oneof=[ODATA, OPEN_API]` |  
format: SpecFormat! | true | `oneof=[YAML, JSON, XML]` |  
fetchRequest: FetchRequestInput | false | |  

### EventDefinitionInput

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` | varchar(256) in db  
description: String | false | `max=128` |  
spec: EventSpecInput! | true | | 
group: String | false | `max=36` | varchar(256) in db  
version: VersionInput | false | |  

### EventSpecInput

- Struct validator ensures that `type` and `format` work together (ASYNC_API works with YAML and JSON)
- Struct validator ensures that only one of `data` and `fetchRequest` is present

Field | Required | Rules | Comment
--- | --- | --- | ---
data: CLOB (string) | false | |  
type: EventSpecType! | true | `oneof=[ASYNC_API]` |  
format: SpecFormat! | true | `oneof=[YAML, JSON]` |  
fetchRequest: FetchRequestInput | false | |  

### VersionInput

Field | Required | Rules | Comment
--- | --- | --- | ---
value: String! | true | `max=256` | varchar(256) in db
deprecated: Boolean = false | true | | required because has default value
deprecatedSince: String | false | `max=256` | varchar(256) in db
forRemoval: Boolean = false | true | | required because has default value

### ApplicationRegisterInput

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` | max 36 characters
description: String | false | `max=128` |  
labels: Labels (map[string]interface{}) | false | key: `required` |  
webhooks: [WebhookInput!] | false | `[required]` |  
healthCheckURL: String | false | `url`, `max=256` | varchar(256) in db  
apiDefinitions: [APIDefinitionInput!] | false | `[required]` |  
eventDefinitions: [EventDefinitionInput!] | false | `[required]` |  
documents: [DocumentInput!] | false | `[required]` |  

### ApplicationUpdateInput

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` | max 36 characters
description: String | false | `max=128` |  
healthCheckURL: String | false | `url`, `max=256` | varchar(256) in db  

### ApplicationTemplateInput

- Struct validator ensures that provided placeholders' names are unique and that they are used in `applicationInput` 

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` |
description: String | false | `max=128` |  
applicationInput: ApplicationCreateInput! | true | |  
placeholders: [PlaceholderDefinitionInput!] | false | `[required]` |  
accessLevel: ApplicationTemplateAccessLevel! | true | `oneof=[GLOBAL]` | 

### PlaceholderDefinitionInput

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` |
description: String | false | `max=128` |

### ApplicationFromTemplateInput

- Struct validator ensures that provided placeholders' names are unique

Field | Required | Rules | Comment
--- | --- | --- | ---
templateName: String! | true | `name` |
values: [TemplateValueInput!] | false | `[required]` |

### TemplateValueInput

Field | Required | Rules | Comment
--- | --- | --- | ---
placeholder: String! | true | `name` |
value: String! | true | | 

### RuntimeInput

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` | varchar(256) in db
description: String | false | `max=128` |
labels: Labels (map[string]interface{}) | false | key: `required` |

### IntegrationSystemInput

Field | Required | Rules | Comment
--- | --- | --- | ---
name: String! | true | `name` | varchar(256) in db  
description: String | false | `max=128` |  

### DocumentInput

Field | Required | Rules | Comment
--- | --- | --- | ---
title: String! | true | `max=128` |  varchar(256) in db
displayName: String! | true | `max=128` |  varchar(256) in db
description: String! | true | `max=128` |  
format: DocumentFormat! | true | `oneof=[MARKDOWN]` |  
kind: String | false | `max=256` |  varchar(256) in db
data: CLOB (string) | false | |  
fetchRequest: FetchRequestInput | false | |  

### WebhookInput

Field | Required | Rules | Comment
--- | --- | --- | ---
type: ApplicationWebhookType! | true | `oneof=[CONFIGURATION_CHANGED]` |
url: String! | true | `url`, `max=256` | varchar(256) in db
auth: AuthInput | false | |

### LabelDefinitionInput

Field | Required | Rules | Comment
--- | --- | --- | ---
key: String! | true | `max=256` | varchar(256) in db  
schema: JSONSchema (string) | false | |  

### LabelInput

Field | Required | Rules | Comment
--- | --- | --- | ---
key: String! | true | `max=256` | varchar(256) in db  
value: Any! (interface{}) | true | | 

### FetchRequestInput

Field | Required | Rules | Comment
--- | --- | --- | ---
url: String! | true | `url`, `max=256` | varchar(256) in db  
auth: AuthInput | false | |  
mode: FetchMode = SINGLE | true | `oneof=[SINGLE, PACKAGE, INDEX]` | required because has default value
filter: String | false | `max=256` | varchar(256) in db  

### AuthInput

Field | Required | Rules | Comment
--- | --- | --- | ---
credential: CredentialDataInput! | true | |  
additionalHeaders: HttpHeaders (map[string][]string) | false | key: `required`, value: `required`, `[required]` |  
additionalQueryParams: QueryParams (map[string][]string) | false | key: `required`, value: `required`, `[required]` |  
requestAuth: CredentialRequestAuthInput | false | | 

### CredentialDataInput

- Struct validator ensuring that exactly one field is not nil

Field | Required | Rules | Comment
--- | --- | --- | ---
basic: BasicCredentialDataInput | false | |  
oauth: OAuthCredentialDataInput | false | |  

### BasicCredentialDataInput

Field | Required | Rules | Comment
--- | --- | --- | ---
username: String! | true | |  
password: String! | true | |  

### OAuthCredentialDataInput

Field | Required | Rules | Comment
--- | --- | --- | ---
clientId: ID! | true | |
clientSecret: String! | true | |
url: String! | true | `url` |

### CredentialRequestAuthInput

- Struct validator ensuring that exactly one field is not nil

Field | Required | Rules | Comment
--- | --- | --- | ---
csrf: CSRFTokenCredentialRequestAuthInput | false | |  

### CSRFTokenCredentialRequestAuthInput

Field | Required | Rules | Comment
--- | --- | --- | ---
tokenEndpointURL: String! | true | `url` |  
credential: CredentialDataInput! | true | | 
additionalHeaders: HttpHeaders (map[string][]string) | false | key: `required`, value: `required`, `[required]` | 
additionalQueryParams: QueryParams (map[string][]string) | false | key: `required`, value: `required`, `[required]` | 
