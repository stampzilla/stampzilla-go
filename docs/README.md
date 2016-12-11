# Documentation

* [Devices](devices.md)


## Architecture

stampzilla-go is composed of alot of standalone services or nodes as we call them (nowdays they are commonly known as microservices).
There is a cli which can be used to start, stop and show status of the running services.
available official nodes can be found in the nodes directory.

All nodes report their state to the server node which have schedule and rule engines. The server also have capabilities to log all state changes as metrics to influxdb
or elasticsearch which mean you can easilly draw temperature graphs or see how long a lamp has been on over time.


## Nodes

* [server](../nodes/stampzilla-server/README.md)




