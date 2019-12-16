# Default eventing for application

## Introduction

The Director as the main component which manages the information for Applications and Runtimes is responsible also for providing eventing configurations. The following document describes the process of determining which runtime should be configured to process the Applications' events.

> **NOTE** Currently only one Runtime can receive events from the Application

Application can obtain the eventing configuration from the Directors API. It is done via the `application(id: ID!): Application`. The `Application` type offers the `eventingConfiguration` property which holds the `defaultURL` property that contains the URL to the eventing backend.
If the value of `defaultURL` is an empty string, it means that there is no Runtime assigned that can process the Application events.

## Determining default Runtime for the Application events

When querying `eventingConfiguration` for the `Application` type the Director component tries to determine the default Runtime that will process the events.
In first step it looks whether there is a Runtime that has following label assigned `{APPLICATION_ID}/defaultEventing = true` and the runtime belongs to the one of the application scenarios. If there is no such runtime, the Director looks for the oldest Runtime that is assigned to the Application scenarios. After having found the oldest Runtime, it is labeled with the label `{APPLICATION_ID}/defaultEventing = true` and next time it will be treated as the default one. Having the default Runtime determined, the Director component fetches its eventing URL and returns as `defaultURL` for the `eventingConfiguration`.

When the Director is unable to find the default Runtime then the `defaultURL` for the `eventingConfiguration` is an empty string.

## Changing the default Runtime assigned for the Application eventing

The Director API offers a mutation that allows assigning the default runtime for the Application eventing. The mutation `setDefaultEventingForApplication(appID: String!, runtimeID: String!): ApplicationEventingConfiguration!` verifies whether the given Runtime belongs to the Application scenarios and labels it with the label `{APPLICATION_ID}/defaultEventing = true`. If the Application had previously assigned default runtime the label from that runtime is removed.

## Deleteing the default Runtime assigned for the Application eventing

THe Directos API offers a moutation that allows deleting the assignment of the default runtime for the Application eventing. The mutation `deleteDefaultEventingForApplication(appID: String!): ApplicationEventingConfiguration!` deletes the label `{APPLICATION_ID}/defaultEventing = true` from the labeled Runtime.

> **NOTE** After deleting the label `{APPLICATION_ID}/defaultEventing = true` and querying for the `eventingConfiguration` for the Application, the Director API will determin the default Runtime for the Application once again.
