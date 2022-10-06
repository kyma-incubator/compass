# Automatic Application Management

> **NOTE**: This document currently contains changes in API that are not yet implemented:
>- Label input in ApplicationRegisterInput


Managing Applications can be passed on some external system, to remove from Applications the burden of integration with Compass.
Because of that [Integration System](../compass/02-01-components.md#Integration-System) was introduced.

To simplify the creation of Applications, Director API is extended with ApplicationTemplate.
An ApplicationTemplate is a reusable input for creating Applications that can be customized with placeholders provided
during Application creation.

## Manage Applications by Integration System

Managing Integrations System is described in [a separate document](./integration-systems.md).
Integration System is uniquely identified by its ID. To register an Application that is managed by an Integration System,
specify `integrationSystemID` in the ApplicationRegisterInput.

```graphql
input ApplicationRegisterInput {
    name: String!
    description: String
    labels: Labels
    webhooks: [WebhookInput!]
    healthCheckURL: String
    apiDefinitions: [APIDefinitionInput!]
    eventDefinitions: [EventDefinitionInput!]
    documents: [DocumentInput!]
    integrationSystemID: ID
}
```
IntegrationSystemID is an optional property. It means that you can still register an Application that is not managed by an IntegrationSystem.

### Example

In this example, Integration System is registered and then it configures newly added Application.

1. Register Integration System

```graphql
mutation {
    registerIntegrationSystem(in: {name: "simpleIntegrationSystem"} ) {
      id
      name
    }
}
```
2. Register Application with specified `integrationSystemID`
```graphql
mutation {
    registerApplication(in:{name:"simpleApplication", integrationSystemID:"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}) {
        id
        name
        integrationSystemID
        labels
    }
}
```
3. Compass adds protected label with name `integrationSystemID` for just registered Application, so output of the previous mutation is the following:
```json
{
  "data": {
    "registerApplication": {
      "id": "d046590f-934f-411f-91e2-d446b404a2a2",
      "name": "simpleApplication",
      "integrationSystemID": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "labels": {
        "scenarios":["DEFAULT"],
        "integrationSystemID":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      }
    }
  }
}
```

Thanks to that,  Integration System can easily fetch all dependant Applications by querying them by label.
Then, Integration System is responsible for updating Application details, like registering API and events definitions.

## Integration System supporting many Application types

In the previous example, there was an assumption that every Application managed by given Integration System represents the same type
of the Application that provides similar API and event definitions.  
In case IntegrationSystem supporting many types of Applications, information about Application type should be stored in the Application labels.
Let assume that IntegrationSystem supports two types of Applications: `ecommerce` and `marketing` and such information
is stored in label `{integration-system-name}/application-type`.
To register an Application of type `ecommerce`, use following mutation:

```graphql
mutation {
    registerApplication(in:{name:"ecommerceApp", integrationSystemID:"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", labels:[{key:"simpleIntegrationSystem/applicationType",value:"ecommerce"}]}) {
        id
        name
        integrationSystemID
        labels
    }
}
```

Drawback of this approach is that label name (`simpleIntegrationSystem/applicationType`) and supported values(`ecommerce`,`marketing`) are arbitrary
defined by a IntegrationSystem and has to be documented.
Luckily, IntegrationSystem can simplify this process by defining ApplicationTemplates to explicitly specify supported types.

## Managing ApplicationTemplates
ApplicationTemplate defines ApplicationInput used to register Application. ApplicationInput can contain a variable part - placeholders.
Placeholders are represented in template in the following form:
```{{placeholder-name}}```
Every placeholder is required. Compass blocks registering Application from template if any placeholder has missing actual value.
In the first iteration ApplicationTemplate will be registered globally and will be visible for all tenants (notice `accessLevel` field)

```graphql
input ApplicationTemplateInput {
    name: String!
    description: String

    applicationInput: ApplicationRegisterInput!
    placeholders: [PlaceholderDefinitionInput!]

    accessLevel: ApplicationTemplateAccessLevel!

}

enum ApplicationTemplateAccessLevel {
    GLOBAL
}


input PlaceholderDefinitionInput {
    name: String!
    description: String
}

input TemplateValueInput {
    placeholder: String!
    value: String!
}

type Mutation {
    createApplicationTemplate(in: ApplicationTemplateInput!): ApplicationTemplate!
    updateApplicationTemplate(id: ID!, in: ApplicationTemplateInput!): ApplicationTemplate!
    deleteApplicationTemplate(id: ID!): ApplicationTemplate!

    registerApplicationFromTemplate(templateName: String!, values: [TemplateValueInput]): Application!

}
type Query {
    applicationTemplates(first: Int = 100, after: PageCursor): ApplicationTemplatePage!
    applicationTemplate(id: ID!): ApplicationTemplate
}
```

### Example

1. Integration System creates Application Template that represents `ecommerce` Application type.

```graphql
mutation {
    createApplicationTemplate(in:{
        name:"ecommerce-template",
        applicationInput:{
            name: "ecommerce-app",
            integrationSystemID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
            labels: [{key:"simpleIntegrationSystem_applicationType",value:"ecommerce"}]
        }
        }) {
           name
    }
 }
```

2. A user lists of all ApplicationTemplates, thanks to that he will find out that there is ApplicationTemplate that represents `ecommerce` Application.

```graphql
query  {
    applicationTemplates {
        data {
            name
            description
        }
    }
}

```
3. A user registers an Application from template:
```graphql
mutation {
    registerApplicationFromTemplate(templateName:"ecommerce-template") {
        id
        name
        labels
    }
}
```

This mutation registers Application with name `ecommerce-app`, integrationSystemID is set to `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa` and
with two labels:
- `integrationSystemID` with value `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa`
- `simpleIntegrationSystem_applicationType` with value `ecommerce`

When Integration System noticed, that new Application was registered, it starts configuring it according to information
stored in the `application-type` label.

In this example, IntegrationSystem registers ApplicationTemplate, but users can also define their own ApplicationTemplates.

## Use placeholders in ApplicationTemplate

In the previous example, ApplicationTemplate was created with hardcoded Application name `ecommerce-app`.
Because Application name has to be unique in given tenant, only one Application from given template can be created.
Fortunately, ApplicationTemplate can define placeholders, that values has to provided when adding an Application.

### Example

In this example we modify previously created ApplicationTemplate and make name configurable.

1. Update ApplicationTemplate

```graphql
mutation {
    updateApplicationTemplate(id:"some-id", in:{
        name:"ecommerce-template",
        applicationInput:{
            name: "{{application-name}}",
            integrationSystemID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
            labels:[{key:"simpleIntegrationSystem_applicationType",value:"ecommerce"}]

        },
        placeholders: [
        {
            name:"application-name",
            description:"Name of the application"
        }],
        }) {
           name
    }
 }
```

As you can see, `application-name` placeholder is defined. In ApplicationInput, we refer to the placeholder in the following form: `{{application-name}}`.

2. Register Application from Template
When user registers Application from Template that defines placeholders, current value for all placeholders has to be specified.

```graphql
mutation {
    registerApplicationFromTemplate(templateName:"ecommerce-template", values: [{placeholder:"application-name", value:"my-aplication"}]) {
        id
        name
        labels
    }
}
```

## Providing input parameters

Application template can also contain a label that defines the JSON schema for the input parameter. Input parameters are sometimes required to configure Application instance by the Integration System. Let assume, that for enabling some Application, the user has to provide credentials that will be used by Integration System to configure the external solution. Such input parameters can be described by appropriate JSON schema and stored under the  `simpleIntegrationSystem/inputParam` label. The Compass does not store credentials itself. It provides only the JSON schema through the API and the consumer is responsible for setting the appropriate values accordingly.

### Example

1. Update the ApplicationTemplate with the label representing input parameters schema used by the Integration System.

```graphql
mutation {
    updateApplicationTemplate(id:"some-id", in:{
        name:"ecommerce-template",
        applicationInput:{
            name: "{{application-name}}",
            integrationSystemID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
            labels:[{key:"simpleIntegrationSystem_applicationType",value:"ecommerce"},
                    {key:"simpleIntegrationSystem_inputParams""{\"type\": \"object\",\"required\": [\"username\",\"password\"],\"properties\": {\"username\": {\"type\": \"string\"},\"password\": {\"type\": \"string\"}}}"}
             ]
        },
        placeholders: [
        {
            name:"application-name",
            description:"Name of the application"
        },
        ],
        }) {
           name
        }
    }
```

2. Register Application
 ```graphql
 mutation {
     registerApplicationFromTemplate(templateName:"ecommerce-template", values: [
     {placeholder:"application-name", value:"MyApplication"}
     ]) {
         id
         name
         labels
    }
}
 ```

## Reasoning

Compass API follows Larry Wall advice:

> Easy things should be easy, and hard things should be possible.

1. User still can register an Application without defining any IntegrationSystem or ApplicationTemplate.
2. ApplicationTemplate can be defined not only by Integration System, but also by users.
If user registers manually many similar Applications, he can define Application Template to simplify it.
3. IntegrationSystem can define ApplicationTemplate, but it is not required. If given IntegrationSystem supports
only one Application type, creating Application template can be an overkill.

From UI perspective, user has also simple view for registering application with two possible options:
- register Application manually
- register Application from Template

# Future plans

1. For every Compass top-level type, it should be possible to define label. Currently, we can add label for Application and Runtime, but in
the future we plan to add possibility to label IntegrationSystem or ApplicationTemplate.
2. To improve customer experience, there should be a possibility to define icon for Application, Runtime and Integration System.
3. Add versioning for ApplicationTemplates. For now, an user can use template name to store information about version.
