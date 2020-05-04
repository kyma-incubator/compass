# Fetch Requests

API Packages support providing documentation for API Definitions. You can provide the documentation
during the call to the Director or use a Fetch Request. Fetch Request is a type in the Director API that contains all the information needed to 
fetch specification from the given URL. It can be provided during the `addPackage`, `addAPIDefinitionToPackage` and `updateAPIDefinition` mutations.
If a Fetch Request is specified, the Director makes a synchronous call to the specified URL and downloads the specification.

You can find information on whether the call is successful in the `Status` field. It contains the condition of the 
Fetch Request, timestamp, and a message stating if something went wrong while fetching the spec. 

If there's an error while fetching eg. wrong URL, the mutation will continue, but an appropriate Fetch Request status will 
be set. 

In case the specification must be fetched again, you can use one of the `refetchSpec` mutations to update the specification.

