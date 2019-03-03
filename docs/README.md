# Documentation

* [Devices](devices.md)


## Architecture

stampzilla-go is composed of alot of standalone services or nodes as we call them (nowdays they are commonly known as microservices).
There is a cli which can be used to start, stop and show status of the running services.
available official nodes can be found in the nodes directory.

All nodes report their state to the server node which have schedule and rule engines. The server also have capabilities to log all state changes as metrics to influxdb which mean you can easilly draw temperature graphs or see how long a lamp has been on over time.


## Nodes

Nodes can be configured in the web interface. The config is a json object and default config for each node can be found in the links below.
Its up to the node if it requires config or not. For example telldus does not require any config but takes all config from the telldusd running on the same machine.

* [deconz](../nodes/stampzilla-deconz/README.md)
* [google-assistant](../nodes/stampzilla-google-assistant/README.md)
* [server](../nodes/stampzilla-server/README.md)
* [telldus](../nodes/stampzilla-telldus/README.md)




