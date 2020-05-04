# Fetch Requests

Packages API supports providing documentation for API Definitions, Event Definitions. User can provide the documentation
during the call to Director or use a Fetch Request. If a Fetch Request is specified Director will call the specified URL
and download the specification.

User can specify these three fields. 

| Name   	| Meaning                                                             	|
|--------	|---------------------------------------------------------------------	|
| Auth   	| Credentials needed to access the url                                	|
| Mode   	| Information if the specifications consists of one or multiple files 	|
| Filter 	| Expresion that specifies which files should be fetched              	|

    >**NOTE:** Currently all of these fields are not supported

Information if the call has been successful, can be found in the `Status` field. It contains the condition of the 
Fetch Request, timestamp, and a message stating if something went wrong.

In case the specification has to be fetched again, user can use one of the `refetchSpec` mutations to update the specification.
