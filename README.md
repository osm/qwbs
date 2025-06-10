# QuakeWorld Broadcast Service

QuakeWorld server whose purpose is to register with the configured master
servers and listen for broadcast messages.

## Configuration

See [qwbs.conf](qwbs.conf) for details on how to configure the service.

## Note

When the service is first launched, it may take some time before it starts
receiving broadcast messages from other servers, as they need to synchronize
with the master servers before those servers know to send messages to us.
