# Restarter

*NOTE: In non-professional state at the moment, I will clean it up and add real
errors, but I wanted to put this up even in its current state*

restarter restarts a command with arguments when a restart channel is sent to.

Current there are two cli interfaces for the library.

## fsrestarter

`fsrestarter` which watches a passed directory for a binary to be updated, it
runs the binary and restarts the binary if it detects any change.

I built this one to be used in dev docker containers. It allows a binary to be
volumed into a container, and then restarted when it is updated. This allows
exposed ports to be entirely within a docker network.

An example `docker-compose.yml` for several services might be:

```
version: "2"
services:
	service-1:
		image: adamryman/fsrestarter
		expose: "5040"
		volumes:
			- $GOPATH/src/github.com/adamryman/service-1/target:/target
		command: target run
		environment:
			PORT: "5040"
			SERVICE-1-HOST: "service-2"
			SERIVCE-1-PORT: "45360"
	service-2:
		image: adamryman/fsrestarter
		expose: "45360"
		volumes:
			- $GOPATH/src/github.com/adamryman/service-2/target:/target
		command: target run
		environment:
			PORT: "45360"
			SERVICE-1-HOST: "service-1"
			SERIVCE-1-PORT: "5040"
	...
```

Where each container will run and try to launch the binary `run` inside the
`target` directory which are volumed together.

## netrestarter

Also `netrestarter` which restarts a passed cmd every time a connection is made
on `localhost:5040` which I made as a test.


