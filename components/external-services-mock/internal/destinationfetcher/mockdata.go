package destinationfetcher

var destinationsSensitiveData = map[string][]byte{
	"s4ext": []byte(`{
      "owner": {
        "SubaccountId": "8fb6ac72-124e-11ed-861d-0242ac120002",
        "InstanceId": null
      },
      "destinationConfiguration": {
        "Name": "s4ext",
        "Type": "HTTP",
        "URL": "https://s4hana.com",
        "Authentication": "BasicAuthentication",
        "ProxyType": "Internet",
        "XFSystemName": "Rock",
        "HTML5.DynamicDestination": "true",
        "User": "usr",
        "product.name": "SAP S/4HANA Cloud",
        "Password": "securePass",
      },
      "authTokens": [
        {
          "type": "Basic",
          "value": "blJhbHQ1==",
          "http_header": {
            "key": "Authorization",
            "value": "Basic blJhbHQ1=="
          }
        }
      ]
    }`),
	"expert": []byte(`
    {
      "owner": {
        "SubaccountId": "66259e9f-b85a-4ecd-a279-486a825f0f8a",
        "InstanceId": null
      },
      "destinationConfiguration": {
        "Name": "expert",
        "Type": "HTTP",
        "URL": "http://test.com",
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
    }`),
}

var destinations = []byte(`[
	{
		"Name": "s4ext",
		"Type": "HTTP"
	},
	{
		"Name": "expert",
		"Type": "HTTP"
	}
]`)
