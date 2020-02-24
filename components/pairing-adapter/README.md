# Pairing Adapter
Pairing Adapter provides REST API that allows you to fetch a one-time token from an External Token Service. The API specification is defined in the `swagger.json` file that is generated from 
code comments with the following command:
```
swagger generate spec -o ./swagger.json cmd/main.go
```

## Configuration

Pairing Adapter binary allows you to override some configuration parameters. You can specify the following environment variables:

| Environment variable                    | Description                                                      |                                                                             
| ----------------------------------------|------------------------------------------------------------------| 
| **MAPPING_TEMPLATE_EXTERNAL_URL**       | External Token Service URL in a form of Golang template that is executed in the context of the `RequestData`.
| **MAPPING_TEMPLATE_HEADERS**            | Headers sent to the External Token Service in a form of Golang template that is executed in the context of the `RequestData`.      
| **MAPPING_TEMPLATE_JSON_BODY**          | Body sent to the External Token Service in a form of Golang template that is executed in the context of the `RequestData`.
| **MAPPING_TEMPLATE_TOKEN_FROM_RESPONSE**| Golang template to get a token from the response received from the External Token Service. 
| **OAUTH_URL**                           | OAuth service URL
| **OAUTH_CLIENT_ID**                     | OAuth client ID
| **OAUTH_CLIENT_SECRET**                 | OAuth client Secret
