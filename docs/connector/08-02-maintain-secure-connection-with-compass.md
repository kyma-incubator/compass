# Maintain a secure connection with Compass

After you have established a secure connection with Compass, you can fetch the configuration details and renew the client certificate before it expires.  
To renew the client certificate, follow the steps in this tutorial.

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- [Established secure connection with Compass](08-01-establish-secure-connection-with-compass.md)
- Compass (version 1.8 or higher)
- Registered Application
- Runtime connected to Compass

> **NOTE**: To see how to establish a secure connection with Compass and generate a client certificate, see [this](08-01-establish-secure-connection-with-compass.md) document. 

<!--- TODO: link in the note above --->

## Steps

1. Get the CSR information with the configuration details.

    To fetch the configuration, make a call to the Certificate-Secured Connector URL using the client certificate. 
    The Certificate-Secured Connector URL is the `certificateSecuredConnectorURL` obtained when establishing a secure connection with Compass. 
    Send this query with the call:
    
    ```graphql
    query {
        result: configuration {
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

    A successful call returns the requested configuration details.

2. Generate a key and a Certificate Signing Request (CSR).

    Generate a CSR with this command using the certificate subject data obtained with the CSR information: 
    ```
    openssl genrsa -out generated.key 2048
    openssl req -new -sha256 -out generated.csr -key generated.key -subj "{SUBJECT}"
    openssl base64 -in generated.csr
    ```

3. Sign the CSR and renew the client certificate. 

    Send this GraphQL mutation with the encoded CSR to the Certificate-Secured Connector URL:

    ```graphql
    mutation {
        result: signCertificateSigningRequest(csr: "{BASE64_ENCODED_CSR}") {
            certificateChain
            caCertificate
            clientCertificate
        }
    }
    ```

    The response contains a renewed client certificate signed by the Kyma Certificate Authority (CA), certificate chain, and the CA certificate.
