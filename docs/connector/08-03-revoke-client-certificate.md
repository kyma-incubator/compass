# Revoke a client certificate

After you have established a secure connection with Compass and generated a client certificate, you may want to revoke this certificate at some point. To revoke a client certificate, follow the steps in this tutorial.

> **NOTE:** A revoked client certificate remains valid until it expires, but it cannot be renewed.

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- Compass (version 1.8 or higher)
- Registered Application
- Runtime connected to Compass
- [Established secure connection with Compass](08-01-establish-secure-connection-with-compass.md)

> **NOTE**: To see how to establish a secure connection with Compass and generate a client certificate, see [this](08-01-establish-secure-connection-with-compass.md) document. To see how to maintain a secure connection with Compass and renew a client certificate, go [here](08-02-maintain-secure-connection-with-compass.md).
<!--- TODO: link in the note above --->

## Steps

1. Revoke the client certificate

    To revoke a client certificate, make a call to the Certificate-Secured Connector URL using the client certificate. 
    The Certificate-Secured Connector URL is the `certificateSecuredConnectorURL` obtained when establishing a secure connection with Compass.
    Send this mutation with the call:
    
    ```graphql
    mutation { result: revokeCertificate }
    ``` 

    A successful call returns the following response:
    
    ```json
    {"data":{"result":true}}
    ```
