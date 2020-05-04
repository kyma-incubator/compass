# Fetch Requests

API Packages support providing documentation for API and Event Definitions. You can provide the documentation
during the call to the Director or use a Fetch Request. If a Fetch Request is specified, the Director calls the specified URL
and downloads the specification.

You can specify the following fields:

| Name   	| Meaning                                                             	|
|--------	|---------------------------------------------------------------------	|
| Auth   	| Credentials needed to access the url                                	|
| Mode   	| Information if the specifications consists of one or multiple files 	|
| Filter 	| Expresion that specifies which files should be fetched              	|

    >**NOTE:** Currently all of these fields are not supported

You can find information on whether the call is successful in the `Status` field. It contains the condition of the 
Fetch Request, timestamp, and a message stating if something went wrong.

In case the specification must be fetched again, you can use one of the `refetchSpec` mutations to update the specification.
