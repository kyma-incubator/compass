# Fetch Requests

Packages API supports providing documentation for API Definitions. You can provide the specification
during the call to the Director or use a Fetch Request. Fetch Request is a type in the Director API that contains all the information needed to 
fetch specification from the given URL. It can be provided during the `addPackage`, `addAPIDefinitionToPackage` and `updateAPIDefinition` mutations.
If a Fetch Request is specified, the Director makes a synchronous call to the specified URL and downloads the specification.

You can find information on whether the call is successful in the `Status` field. It contains the condition of the 
Fetch Request, a timestamp, and a message that states whether something went wrong when fetching the specification. 

If there is an error while fetching the specification, the mutation is continued, but an appropriate Fetch Request status is set. 

In case the specification must be fetched again, you can use one of the `refetchSpec` mutations to update the specification.
