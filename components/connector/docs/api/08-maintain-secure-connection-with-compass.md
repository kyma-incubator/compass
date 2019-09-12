---
title: Maintain a secure connection with Compass
type: Tutorials
---

After you have established a secure connection with Compass, you can fetch the configuration details and renew the client certificate before it expires.  
This guide shows you how to do that. 
<!--- TODO ---> 

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- Established secure connection with Compass

> **NOTE**: To see how to establish a secure connection with Compass, see [this]() document. 

<!--- TODO: link in the note above --->

## Get the CSR information and configuration details from Kyma using the client certificate 

To fetch the configuration, make a call to the Certificate-Secured Connector URL using the client certificate. 
The Certificate-Secured Connector URL is the `certificateSecuredConnectorURL` obtained when establishing a secure connection with Compass. 
Send this query with the call:
```
query {
    result: configuration() {
        certificateSigningRequestInfo { 
            subject 
            keyAlgorithm 
        }
        managementPlaneInfo { 
            directorURL 
        }
    }
}
``` 

A successful call returns the data requested in the query.

## Renew the client certificate 

Generate a CSR with this command using the certificate subject data obtained with the CRS information: 
```
openssl genrsa -out generated.key 2048
openssl req -new -sha256 -out generated.csr -key generated.key -subj "{SUBJECT}"
openssl base64 -in generated.csr
```

Replace the `BASE64_ENCODED_CSR` with the encoded CSR in this GraphQL mutation:

```
mutation {
    result: signCertificateSigningRequest(csr: "BASE64_ENCODED_CSR") {
        certificateChain
        caCertificate
        clientCertificate
    }
}
```
Send the modified GraphQL mutation to the Certificate-Secured Connector URL.

The response contains a renewed client certificate signed by the Kyma Certificate Authority (CA) and the CA certificate.

After you receive the certificate, decode it and use it in your application. 
