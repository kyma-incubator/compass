# Automatic Application Management

Managing Applications can be passed on some external system, to remove from Applications the burden of integration with Compass.
Because of that [Integration System](./components.md#Integration-System) was introduced.

To simplify creation of Applications, Director API is extended with ApplicationTemplate.
An ApplicationTemplate is reusable input for creating Applications that can be customized with placeholders provided
during Application creation.

## Manage Applications by Integration System
Managing Integrations System is described in [a separate document](./integration-systems.md).
Integration System is uniquely identified by its ID. To create an Application that is managed by an Integration System,
specify `integrationSystemID` in the ApplicationCreateInput.

```graphql
input ApplicationCreateInput {
    name: String!
    description: String
    labels: Labels
    webhooks: [WebhookInput!]
    healthCheckURL: String
    apis: [APIDefinitionInput!]
    eventAPIs: [EventAPIDefinitionInput!]
    documents: [DocumentInput!]
    integrationSystemID: ID
    labels: [LabelInput!]
}
```
IntegrationSystemID is an optional property. It means that you can still create Application that is not managed by IntegrationSystem.

### Example
In this example, Integration System is registered and then it configures newly added Application.
1. Register Integration System
```graphql
mutation {
    createIntegrationSystem(in: {name: "simpleIntegrationSystem"} ) {
      id
      name
    }
}

```
2. Create Application with specified `integrationSystemID`
```graphql
mutation {
    createApplication(in:{name:"simpleApplication", integrationSystemID:"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}) {
        id
        name
        integrationSystemID
        labels
    }
}
```
Compass adds protected label with name `integrationSystemID` for just created Application, so output of the previous mutation is the following:
```json
{
  "data": {
    "createApplication": {
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
Then, Integration System is responsible for updating Application details, like registering API and events.
Given Integration System has privileges for modifying only Applications with matching `integrationSystemID`, so it cannot
modify Applications managed manually or managed by other Integration System.

## Integration System supporting many Application types
In the previous example, there was an assumption that every Application managed by given Integration System represents the same type
of the Application that provides similar API and event definitions.  
In case IntegrationSystem supporting many types of Applications, information about Application type should be stored in the Application labels.
Let assume that IntegrationSystem supports two types of Applications: `ecommerce` and `marketing` and such information
is stored in label `{integration-system-name}/applicationType`.
To create an Application of type `ecommerce`, use following mutation:
```graphql
mutation {
    createApplication(in:{name:"ecommerceApp", integrationSystemID:"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", labels:[{key:"simpleIntegrationSystem/applicationType",value:"ecommerce"}]}) {
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
ApplicationTemplate defines ApplicationInput used to create Application. ApplicationInput can contains variable part - placeholders.
Placeholders are represented in template in the following form:
```{{PLACEHOLDER_NAME}}```
Every placeholder is required. Compass blocks creating Application from template if any placeholder has missing actual value.
In the first iteration ApplicationTemplate will be registered globally and will be visible for all tenants (notice `accessLevel` field)

```graphql
input ApplicationTemplateInput {
    name: String!
    description: String

    applicationInput: ApplicationCreateInput!
    placeholders:       [PlaceholderDefinitionInput!]

    accessLevel: ApplicationTemplateAccessLevel!

}

enum ApplicationTemplateAccessLevel {
    GLOBAL
}


input PlaceholderDefinitionInput {
    name            String!
    description     String
}

input TemplateValueInput {
    placeholder     String!
    value           String!
}

type Mutation {
    createApplicationTemplate(in: ApplicationTemplateInput!): ApplicationTemplate!
    updateApplicationTemplate(id: ID!, in: ApplicationTemplateInput!): ApplicationTemplate!
    deleteApplicationTemplate(id: ID!): ApplicationTemplate!

    createApplicationFromTemplate(templateName: String!, values: [TemplateValueInput]): Application!

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
            labels:[{key:"simpleIntegrationSystem/applicationType",value:"ecommerce"}]
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


2. A user creates an Application from template:
```graphql
mutation {
    createApplicationFromTemplate(templateName:"ecommerce-template") {
        id
        name
        labels
}
```

This mutation creates Application with name `ecommerce-app`, integrationSystemID is set to `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa` and
with two labels:
- `integrationSystemID` with value `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa`
- `simpleIntegrationSystem/applicationType` with value `ecommerce`

When Integration System noticed, that new Application was created, it starts configuring it according to information
stored in the `applicationType` label.

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
            name: "{{APPLICATION_NAME}}",
            integrationSystemID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
            labels:[{key:"simpleIntegrationSystem/applicationType",value:"ecommerce"}]

        },
        placeholders: [
        {
            name:"APPLICATION_NAME",
            description:"Name of the application"
        }],
        }) {
           name
    }
 }
```

As you can see, `APPLICATION_NAME` placeholder is defined. In ApplicationInput, we refer to the placeholder in the following form: `{{APPLICATION_NAME}}`.

2. Create Application from Template
When user creates Application from Template that defines placeholders, current value for all placeholders has to be specified.

```graphql
mutation {
    createApplicationFromTemplate(templateName:"ecommerce-template", values: [{placeholder:"APPLICATION_NAME", value:"MyApplication"}]) {
        id
        name
        labels
}
```

## Use placeholders for providing Input Parameters

Placeholders can be used also for providing input parameters required for configuring external Applications by Integration System.
Let assume, that for enabling some Application, user has to provide credentials that will be used by Integration System
to configure external solution. Such credentials can be stored in the labels.
Because labels can store credentials, Compass has to ensure that given label can be read only by specific Integration System.

### Example
1. Update ApplicationTemplate with labels representing input parameters used by Integration System
```graphql
mutation {
    updateApplicationTemplate(id:"some-id", in:{
        name:"ecommerce-template",
        applicationInput:{
            name: "{{APPLICATION_NAME}}",
            integrationSystemID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
            labels:[{key:"simpleIntegrationSystem/applicationType",value:"ecommerce"},
                    {key:"simpleIntegrationSystem/inputParam/username", value:"{{USERNAME}}"},
                     {key:"simpleIntegrationSystem/inputParam/password", value:"{{PASSWORD}}"},
             ]

        },
        placeholders: [
        {
            name:"APPLICATION_NAME",
            description:"Name of the application"
        },
        {
            name:"USERNAME",
            description:"User name"
        },
        {
            name:"PASSWORD",
            description:"Password"
        },

        ],
        }) {
           name
    }
 }
```

2. Create Application
 ```graphql
 mutation {
     createApplicationFromTemplate(templateName:"ecommerce-template", values: [
     {placeholder:"APPLICATION_NAME", value:"MyApplication"},
     {placeholder:"USERNAME", value:"john@doe.com"},
     {placeholder:"PASSWORD", value:"perch"}
     ]) {
         id
         name
         labels
 }
 ```

## Reasoning
Compass API follows Larry Wall advice:
> Easy things should be easy, and hard things should be possible.

1. User still can create an Application without defining any IntegrationSystem or ApplicationTemplate.
2. ApplicationTemplate can be defined not only by Integration System, but also by users.
If user creates manually many similar Applications, he can define Application Template to simplify it.
3. IntegrationSystem can define ApplicationTemplate, but it is not required. If given IntegrationSystem supports
only one Application type, creating Application template can be an overkill.

From UI perspective, user has also simple view for creating application with two possible options:
- create manually Application
- create Application from Template

# Future plans
1. For every Compass top-level type, it should be possible to define label. Currently, we can add label for Application and Runtime, but in
the future we plan to add possibility to label IntegrationSystem or ApplicationTemplate.
2. To improve customer experience, there should be a possibility to define icon for Application, Runtime and Integration System.
3. Add versioning for ApplicationTemplates. For now, an user can use template name to store information about version.
