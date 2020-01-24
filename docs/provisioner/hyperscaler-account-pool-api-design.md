# Hyperscaler Account Pool API design

## Introduction

The goal of this document is to propose design for Hyperscaller Account Pool API.

## Credentials data needed for provisioning clusters with Hydroform

### Gardener

The following fields must be returned from HAP:

- Gardener project name
- Target provider secret name in Gardener
- Kubeconfig

### GCP

The following fields must be returned from HAP:

- Project name
- Service account key

### AWS

Not defined yet. Design of the API should allow enhancement. 

### Azure

Not defined yet. Design of the API should allow enhancement. 

## Input data to be used by HAP

The following data need to be provided by the API user in Cockpit scenario:

- what type of Hyperscaler to use: e.g. GCP
- Account name (extracted from header)
- Credential Name

The following data need to be provided by the API user in SCP scenario:

- what type of Hyperscaler to use: e.g. Gardener-Azure

- Account name (extracted from Header)

- Credential Name : it should be discussed whether Environment Broker will have enough knowledge to pass credential name to the Provisioner ; there is a question whether HAP should be able to determine what is the Target Provider Secret Name 

   

  ## Possible solutions

  ### Proposal #1 - single GraphQL mutation returning union type

  ```
  input HyperscalerInput {
      Hyperscaler: String
      Account: String
      CredentialName: String  # Should it be provided in Gardener scenario?
      PoolName: String
  }
  
  type GardenerCredentials {
      ProjectName: String
      TargetProviderSecretName: String
      Kubeconfig: String
  }
  
  type GCPCredentials {
      ProjectName: String
      ServiceAccountKey: String
  }
  
  union Credentials = GardenerCredentials | GCPCredentials
  
  type Mutation {
      GetCredentials(input: HyperscalerInput): Credentials
  }
  ```

  Pros:

  - Simplicity - one method only

  Cons:

  - Problems with handling unions in Golang client ; the response needs to be unmarshalled to some concrete type or to a map - it results in ugly code

  

  ### Proposal #2 - single GraphQL mutation returning object 

  ```
  input HyperscalerInput {
      Hyperscaler: String!
      Account: String!
      CredentialName: String 
      PoolName: String!
  }
  
  type GardenerCredentials {
      ProjectName: String
      TargetProviderSecretName: String
      Kubeconfig: String
  }
  
  type GCPCredentials {
      ProjectName: String
      ServiceAccountKey: String
  }
  
  type Credentials {
      GardenerCredentials: GardenerCredentials
      GCPCredentials: GCPCredentials
  }
  
  type Mutation {
      GetCredentials(input: HyperscalerInput): Credentials
  }
  ```

  Pros:

  - Simplicity - one method only

    

  ### Proposal #3 - separate GraphQL mutations

  ```
  input HyperscalerInput {
      Account: String!
      CredentialName: String
      PoolName: String!
  }
  
  type GardenerCredentials {
      ProjectName: String
      TargetProviderSecretName: String
      Kubeconfig: String
  }
  
  type GCPCredentials {
      ProjectName: String # How to obtain Project Name in Cockpit scenario
      ServiceAccountKey: String
  }
  
  type Mutation {
      GetGardenerCredentials(input: HyperscalerInput, provider: String): GardenerCredentials
      GetGCPCredentials(input: HyperscalerInput) : GCPCredentials
  }
  ```

  Cons:

  - Two methods to be tested

  

  ### Design decision

  The second proposal seems to be most suitable.

  

  

  

  

  

  

  

  

  