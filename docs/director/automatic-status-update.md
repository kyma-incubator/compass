# Automatic status update

Applications and Runtimes in Director use the `Status` field to inform the user about their condition. The field can be 
set externally by updating the object, but it is also automatically set by the Director.
Every time an Application or Runtime calls the API, it's status is automatically set to `CONNECTED` regardless of the previous one.