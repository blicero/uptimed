# uptimed

uptimed is a simple application to keep track of the uptime and system
load of Unix-like systems.

The server side provides a HTTP interface that is both used by the
clients to report data and to provide a user interface.

The client side gathers and reports data about the uptime and system
load of the system it runs on and sends that data to the server. If
the server is not reachable, data is saved locally and transmitted as
soon as the server becomes available again.

See the [Journal](journal.org) for details about the internals and the
progress of development over time.
