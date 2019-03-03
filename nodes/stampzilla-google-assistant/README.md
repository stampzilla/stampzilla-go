# google-assistant

Exposes all the devices in the server for control in google assistant (google home etc)

## Configuration

The port is what port to listen to for google actions API calls. That port must be avilable from the internet.
You need to create a project in google actions console: https://developers.google.com/actions/smarthome/create#create-project
And fill in the values in the config below.

```
{
	"port": "8000",
	"clientID": "",
	"clientSecret": "",
	"projectID": "",
	"APIKey": ""
}
```
