# Establish a secure connection with Compass

To establish a secure connection with Compass and generate the client certificate, follow this tutorial. 

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- Compass (version 1.8 or higher)
- Registered Application
- Runtime connected to Compass

## Steps

1. Get the Connector URL and the one-time token.

    To get the Connector URL and the one-time token which allow you to fetch the required configuration details, use the Compass Console.
    
    Alternatively, make a call to the Director including the `Tenant` header with Tenant ID and `authorization` header with the Dex Bearer token. Use the following mutation: 
    
    ```graphql
    mutation { 
        result: generateOneTimeTokenForApplication(id: "{APPLICATION_ID}") { 
            token 
            connectorURL 
        }
    }
    ```
   
   > **NOTE:** The one-time token expires after 5 minutes.

2. Get the CSR information and configuration details from Kyma using the one-time token.

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

    A successful call returns the data requested in the query including a new one-time token.

3. Generate a key and a Certificate Signing Request (CSR).

    Generate a CSR with this command using the certificate subject data obtained with the CSR information: 
    
    ```bash
    openssl genrsa -out generated.key 2048
    openssl req -new -sha256 -out generated.csr -key generated.key -subj "{SUBJECT}"
    openssl base64 -in generated.csr
    ```

4. Sign the CSR and get a client certificate. 

    To get the CSR signed, use the encoded CSR in this GraphQL mutation:
    
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

    The response contains a certificate chain, a valid client certificate signed by the Kyma Certificate Authority (CA), and the CA certificate.
    
    After you receive the certificates, decode the certificate chain with the base64 method and use it in your application. 
    
 >**NOTE:** To learn how to renew a client certificate, read [this](08-02-maintain-secure-connection-with-compass.md) document.