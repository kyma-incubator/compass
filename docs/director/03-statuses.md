# Application and Runtime status

Applications and Runtimes in the Director use the `Status` field to inform the user about their condition. The possible conditions are:
- `INITIAL` - used for newly created Applications and Runtimes that have not performed any calls to the Director yet
- `PROVISIONING` - used for Runtimes that are being provisioned
- `CONNECTED` - used for connected Applications and Runtimes
- `FAILED` - used for Applications and Runtimes whose connection with the Director failed

## Automatic status update

When a given Application or Runtime communicates with the Director API for the first time, the Director automatically sets the entity's status to `CONNECTED`.

The communication can be either direct or through the Integration System.
- If the Integration System registers Applications and Runtimes in the Director, it manages the statuses of Runtimes and Applications on its own.
- If you register Applications and Runtimes directly in the Director, you can manage the statuses manually.

In both ways of communication, the statuses of Applications and Runtimes are automatically set to `CONNECTED` every time they connect with the Director API, regardless of their previous condition.
