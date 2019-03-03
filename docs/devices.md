# Devices

A device is a single unit which is controllable and/or reports some data. For example sensor, light etc.

### Device types

type | Description 
--- | --- 
light | Can switch on or off 
sensor | can report any sensor data for example temperature
button | Is a momentary button which can be pressed


### Traits

A device can have one or multiple traits. This was inspired by the google actions API. 

##### OnOff

Required states

state | type 
--- | --- 
on | bool

##### Brightness

Required states

state | type 
--- | --- 
brightness | float 0-1

##### ColorSetting

Required states

state | type 
--- | --- 
temperature | int 2000-6500 kelvin

