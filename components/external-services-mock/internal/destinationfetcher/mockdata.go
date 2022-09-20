package destinationfetcher

var sensitiveDataTemplate = `
    {
      "owner": {
        "SubaccountId": "%s",
        "InstanceId": null
      },
      "destinationConfiguration": {
        "Name": "%s",
        "Type": "%s",
        "URL": "http://%s.com",
        "Authentication": "BasicAuthentication",
        "ProxyType": "Internet",
        "User": "usr",
        "Password": "pass"
      },
      "authTokens": [
        {
          "type": "Basic",
          "value": "SkQyZ",
          "http_header": {
            "key": "Authorization",
            "value": "Basic SkQyZ"
          }
        }
      ]
    }`
