# stampzilla-huebridge


### Installation
```
go get github.com/stampzilla/stampzilla-go/nodes/stampzilla-huebridge
```


### Run standalone without stampzilla-server and only call URLs when activated
```
stampzilla-huebridge -standalone
```


Note that google home only supports port 80 for huebridge so make sure you allow it

in linux you can do it like this:
```
setcap cap_net_bind_service=+ep $GOPATH/bin/stampzilla-huebridge
```


### Example configuration when using standalone mode

Save as config.json in current working directory
```json
{
	"Node": {
		"ListenPort": "80",
		"Port": "8080",
		"Devices": [
			{
				"name": "Test1",
				"id": 1,
				"url": {
					"level": "http://localhost/dim/%f",
					"on": "http://localhost/on",
					"off": "http://localhost/off"
				}
			},
			{
				"name": "Test2",
				"id": 2,
				"url": {
					"level": "http://localhost/dim/%f",
					"on": "http://localhost/on",
					"off": "http://localhost/off"
				}
			}
		]
	}
}

```
