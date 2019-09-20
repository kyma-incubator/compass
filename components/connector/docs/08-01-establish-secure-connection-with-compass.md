---
title: Establish a secure connection with Compass
type: Tutorials
---

To establish a secure connection with Compass and generate the client certificate, follow this tutorial. 

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards

## Get the Connector URL and the one-time token

To get the Connector URL and the one-time token which allow you to fetch the required configuration details, use the Compass Console.
<!--- TODO --->

## Get the CSR information and configuration details from Kyma using the one-time token

To get the CSR information and configuration details, send this GraphQL query with the one-time token included in the `connector-token` header to the Connector URL:

```graphql
query {
    result: configuration {
        token {
            token
        }
        certificateSigningRequestInfo {
            subject
            keyAlgorithm
        }
        managementPlaneInfo {
            directorURL
            certificateSecuredConnectorURL
        }
    }
}
``` 

A successful call returns the data requested in the query and a new one-time token.

## Generate the client certificate

Generate a CSR with this command using the certificate subject data obtained with the CSR information: 
```
openssl genrsa -out generated.key 2048
openssl req -new -sha256 -out generated.csr -key generated.key -subj "{SUBJECT}"
openssl base64 -in generated.csr
```

Use the encoded CSR in this GraphQL mutation:
```graphql
mutation {
    result: signCertificateSigningRequest(csr: "{BASE64_ENCODED_CSR}") {
        certificateChain
        caCertificate
        clientCertificate
    }
}
```
Send the modified GraphQL mutation to the Connector URL including the one-time token fetched with the configuration in the `connector-token` header.

The response contains a valid client certificate signed by the Kyma Certificate Authority (CA) and the CA certificate.

After you receive the certificate, decode it with the base64 method and use it in your application. 