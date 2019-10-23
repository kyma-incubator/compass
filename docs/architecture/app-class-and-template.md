# ApplicationClass and ApplicationTemplate

Managing [Applications](./../terminology.md#Application) can be passed on some external system, to remove the burden of integration with Compass from Applications.
Because of that [IntegrationSystems](./../terminology.md#Integration-System) were introduced.

To simplify creation of Applications, Director API is extended with Application Template. 
An ApplicationTemplate is reusable input for creating Applications that can be customized with placeholders provided 
during Application creation.

## Manage Application by Integration System
Managing Integrations System is described in [a separate document](./integration-systems.md).
Integration System can be uniquely identified by it's name. To create an Application that is managed by a Integration System,
specify `integrationSystemName` in the ApplicationCreateInput. 

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
    integrationSystemName: String #optional
    labels: [LabelInput!]
}
```
IntegrationSystemID is an optional. 
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
2. Create Application with provided `integrationSystemName`
```graphql
mutation {
    createApplication(in:{name:"simpleApplication", integrationSystemName:"simpleIntegrationSystem"}) {
        id
        name
        integrationSystemName
        labels
    }
}
```
Compass add labels with integrationSystemName for just created Application, so output of the previous mutation is the following:
```
{
  "data": {
    "createApplication": {
      "id": "d046590f-934f-411f-91e2-d446b404a2a2",
      "name": "simpleApplication",
      "integrationSystemName": "simpleIntegrationSystem",
      "labels": {
        "scenarios":["DEFULT"],
        "integrationSystemName":"simpleIntegrationSystem",
      },
      
    }
  }
}
```

Thanks to that,  Integration System can easily fetch all Applications managed by given Integration System by querying by label.
Then, Integration System is responsible for updating Application details, like registerting API and events.

## Integration System supporting many Application types
In the previous example, there was an assumption that every Application managed by the Integration System represents the same type
of the Application.  
In case IntegrationSystem supports many types of Applications, information about which Application type should be stored in the Application labels.
Let assume that IntegrationSystem supports two types of Applications `ecommerce` and `marketing` and such information 
is stored in label `simpleIntegrationSystem/applicationType`.
To create an Application of type `ecommerce`, create following mutation:
```graphql
mutation {
    createApplication(in:{name:"ecommerceApp", integrationSystemName:"simpleIntegrationSystem", labels:[{key:"simpleIntegrationSystem/applicationType",value:"ecommerce"}]}) {
        id
        name
        integrationSystemName
        labels
    }
}
```
Label name (`simpleIntegrationSystem/applicationType`) and supported values(`ecommerce`,`marketing`) are arbitrary 
defined by a IntegrationSystem and has to be documented.
IntegrationSystem can use ApplicationTemplates to explicitly define supported types.

## Managing ApplicationTemplates
ApplicationTemplate defines ApplicationInput used to create Application. ApplicationInput can contains variable part - placeholders.

```graphql
input ApplicationTemplateInput {
    name: String!
    description: String

    applicationInput: ApplicationCreateInput!
    placeholders:       [PlaceholderDefinitionInput!]

}

input PlaceholderDefinitionInput {
    Name            String!
    Description     String
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
    createAppl
}

```


    
 
## Later
- everything is labelled, including integration system
- application input can define icon
- versioning of application class


describe UI:
- 2 links: create Application and create Application from Template

- security