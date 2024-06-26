package destinationsvc

var FindAPINoAuthDestResponseTemplate = `
{
  "owner": {
    "SubaccountId": "%s",
    "InstanceId": "%s"
  },
  "destinationConfiguration": {
    "Name": "%s",
    "Type": "%s",
    "URL": "%s",
    "AuthenticationType": "%s",
	"Authentication": "%s",
    "ProxyType": "%s"
  }
}`

var FindAPIBasicDestResponseTemplate = `
{
  "owner": {
    "SubaccountId": "%s",
    "InstanceId": "%s"
  },
  "destinationConfiguration": {
    "Name": "%s",
    "Type": "%s",
    "URL": "%s",
    "AuthenticationType": "%s",
	"Authentication": "%s",
    "ProxyType": "%s",
    "User": "%s",
    "Password": "%s"
  },
  "authTokens": [
    {
      "type": "Basic",
      "value": "bXktZmlyc3QtdXNlcjpzZWNyZXRQYXNzd29yZA==",
      "http_header": {
        "key": "Authorization",
        "value": "Basic bXktZmlyc3QtdXNlcjpzZWNyZXRQYXNzd29yZA=="
      }
    }
  ]
}`

var FindAPISAMLAssertionDestResponseTemplate = `
{
  "owner": {
    "SubaccountId": "%s",
    "InstanceId": "%s"
  },
  "destinationConfiguration": {
    "Name": "%s",
    "Type": "%s",
    "URL": "%s",
    "AuthenticationType": "%s",
	"Authentication": "%s",
    "ProxyType": "%s",
    "audience": "%s",
    "KeyStoreLocation": "%s"
  },
  "certificates": [
    {
      "Name": "%s",
      "Content": "/u3+7QA"
    }
  ],
  "authTokens": [
    {
      "type": "SAML2.0",
      "value": "PD94bW",
      "http_header": {
        "key": "Authorization",
        "value": "SAML2.0 PD94bW"
      }
    }
  ]
}`

var FindAPIClientCertDestResponseTemplate = `
{
  "owner": {
    "SubaccountId": "%s",
    "InstanceId": "%s"
  },
  "destinationConfiguration": {
    "Name": "%s",
    "Type": "%s",
    "URL": "%s",
    "AuthenticationType": "%s",
	"Authentication": "%s",
    "ProxyType": "%s",
    "KeyStoreLocation": "%s"
  },
  "certificates": [
    {
      "Name": "%s",
      "Content": "/u3+7QA"
    }
  ]
}`

var FindAPIOAuth2ClientCredsDestResponseTemplate = `
{
  "owner": {
    "SubaccountId": "%s",
    "InstanceId": "%s"
  },
  "destinationConfiguration": {
    "Name": "%s",
    "Type": "%s",
    "URL": "%s",
    "AuthenticationType": "%s",
	"Authentication": "%s",
    "ProxyType": "%s",
    "clientId": "%s",
    "clientSecret": "%s",
    "tokenServiceURL": "%s"
  },
  "authTokens": [
    {
      "type": "bearer",
      "value": "eyJhbGc",
      "http_header": {
        "key": "Authorization",
        "value": "Bearer eyJhbGc"
      },
      "expires_in": "43199",
      "scope": "test.scope"
    }
  ]
}`

var FindAPIOAuth2mTLSDestResponseTemplate = `
{
    "owner": {
        "SubaccountId": "%s",
        "InstanceId": "%s"
    },
    "destinationConfiguration": {
        "Name": "%s",
        "Type": "%s",
        "URL": "%s",
        "AuthenticationType": "%s",
		"Authentication": "%s",
        "ProxyType": "%s",
        "tokenServiceURLType": "%s",
        "clientId": "%s",
        "tokenServiceURL": "%s",
        "tokenService.KeyStoreLocation": "%s"
    },
    "authTokens": [
        {
            "type": "bearer",
            "value": "eyJhbGc",
            "http_header": {
                "key": "Authorization",
                "value": "Bearer eyJhbGc"
            },
            "expires_in": "43199",
            "scope": "test.scope"
        }
    ]
}`
