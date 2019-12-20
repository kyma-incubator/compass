---
title: Connector
type: Architecture
---

The Connector generates client certificates which are used to secure the communication between Compass and the connected external solutions.        

Generating a new client certificate is the first step in the process of configuring an Application. Kyma stores the Root Certificate and serves as the Certificate Authority (CA) when you configure a new Application. When you generate a new client certificate, the Connectors returns it along with the CA certificate (the Root Certificate) to allow validation.  

This diagram illustrates the client certificate generation flow in details:

![Client certificate generation operation flow](assets/001-connection-flow.svg)

1. The administrator requests a token using <!--- the CLI or --> the UI and receives the Connector URL and the one-time token, which is valid for a limited period of time.
    >**NOTE:** The Director creates and stores mapping data, which is used to map the Application ID onto the Tenant ID. 
2. The administrator passes the token and the connection URL to the external system, which requests information regarding the Kyma configuration and CSR information. In the response, it receives:
    - the certificate-secured Connector URL
    - the URL of the Compass Director
    - information required to generate a CSR
    - new one-time token
3. The external system generates a CSR based on the information provided by Kyma and sends the CSR back to the Connector URL. In the response, the external system receives a signed certificate and the CA certificate. It can use the certificate to authenticate and safely communicate with Kyma.

The external application must not hardcode any URLs. The external application must store the Connector URL, the Director URL and the certificate-secured Connector URL along with the certificate. 

The external application can fetch configuration information using the client certificate. It uses this information to generate a CSR prior to certificate renewal. This approach makes certificate rotation process convenient and flexible, since the external application does not need to store information required to generate a CSR in its data model.     

>**NOTE:** To establish a secure connection, follow [this](08-01-establish-secure-connection-with-compass.md) guide.  
> To mainatain a secure connection, see [this](08-02-maintain-secure-connection-with-compass.md) tutorial.