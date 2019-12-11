---
title: Connector Service
type: Architecture
---

The Connector Service generates client certificates which are used to secure the communication between Compass and the connected external solutions.        

Generating a new client certificate is the first step in the process of configuring an Application (App). Kyma stores the root certificate and serves as the Certificate Authority (CA) when you configure a new App. When you generate a new client certificate, the Connector Service returns it along with the CA certificate (the root certificate) to allow validation.  

This diagram illustrates the client certificate generation flow in details:

![Client certificate generation operation flow](assets/001-connection-flow.svg)

1. The administrator requests a token using <!--- the CLI or --> the UI and receives the Connector URL and the one-time token, which is valid for a limited period of time.
2. The administrator passes the token and the connection URL to the external system, which requests information regarding the Kyma configuration and CSR information. In the response, it receives:
    - the certificate-secured Connector URL
    - the URL of the Compass Director
    - information required to generate a CSR
    - new one-time token
3. The external system generates a CSR based on the information provided by Kyma and sends the CSR back to the Connector URL. In the response, the external system receives a signed certificate and the CA certificate. It can use the certificate to authenticate and safely communicate with Kyma.

>**NOTE:** The external application must not hardcode any URLs. The external application must store the Connector URL, the Director URL and the certificate-secured Connector URL along with the certificate. 

>**NOTE:**  The external application can fetch configuration information using the client certificate. It uses this information to generate a CSR prior to certificate renewal. This approach makes certificate rotation process convenient and flexible, since the external application does not need to store information required to generate a CSR in its data model.     

>**NOTE:** To establish a secure connection, follow [this](../../../docs/tutorials/08-01-establish-secure-connection-with-compass.md) guide.  
> To mainatain a secure connection, see [this](../../../docs/tutorials/08-02-maintain-secure-connection-with-compass.md) tutorial.