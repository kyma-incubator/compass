# Packages API

## Introduction

On Kyma Runtime, every Application is represented as a single Service Class, and every Package of a given Application is represented as a single Service Plan. 

![API Packages Diagram](./assets/packages-api.svg)

This document describes API for Packages, which groups multiple API Definitions, Event Definitions and Documents.

## Assumptions
- A single API or Event Definition can be a part of a single Package.
- Package belongs to a single Application entity.

## GraphQL API

In order to manage Packages, Director exposes the following GraphQL API:

```graphql

```

## Package credentials

To read about Package credentials flow, API, how to provide optional input parameters during Service Instance creation, see the [Credential requests for Packages](./03-packages-credential-requests.md) document.
