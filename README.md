# Restarter

*NOTE: In non-professional state at the moment, I will clean it up and add real
errors, but I wanted to put this up even in its current state*

restarter restarts a command with arguments when a restart channel is sent to.

Current there are two cli interfaces for the library.

`fsrestarter` which watches a passed directory for a binary to be updated, it
runs the binary and restarts the binary if it detects any change.

I built this one to be used in dev docker containers. It allows a binary to be
volumed into a container, and then restarted when it is updated. This allows
exposed ports to be entirely within a docker network.

Also `netrestarter` which restarts a passed cmd every time a connection is made
on `localhost:5040` which I made as a test.


