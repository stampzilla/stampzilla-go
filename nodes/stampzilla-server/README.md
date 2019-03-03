# Server

The server is the main component. It has the following responsibilities among other things.

* Stores the config for all the nodes and sent it to them when they start.
* Serves a web gui.
* Stores and evalutate rules. 
* Stores and runs schedules.


### Developing

Install deps

```
dep ensure -vendor-only

```

Allow https to localhost in chrome
```
chrome://flags/#allow-insecure-localhost
```
