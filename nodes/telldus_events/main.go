package main

import (
    "fmt"
    "unsafe"
)

/*
#cgo LDFLAGS: -ltelldus-core

#include <telldus-core.h>

extern void registerCallbacks();
extern void unregisterCallbacks();

*/
import "C"


func main() {
    fmt.Println("Startup...");

    C.registerCallbacks();
    defer C.unregisterCallbacks();

    select{}
}

//export sensorEvent
func sensorEvent(protocol, model *C.char, sensorId, dataType int, value *C.char) {
    fmt.Printf("SensorEVENT %s,\t%s,\t%d -> ", C.GoString(protocol), C.GoString(model), sensorId);

	if (dataType == C.TELLSTICK_TEMPERATURE) {
		fmt.Printf("Temperature:\t%s\n", C.GoString(value));
		
	} else if (dataType == C.TELLSTICK_HUMIDITY) {
		fmt.Printf("Humidity:\t%s%%\n", C.GoString(value));
	}
}


//export deviceEvent
func deviceEvent(deviceId,method int,data *C.char,callbackId int,context unsafe.Pointer) {
    fmt.Printf("DeviceEVENT %d\t%d\t:%s\n",deviceId,method,C.GoString(data));
}

//export deviceChangeEvent
func deviceChangeEvent(deviceId, changeEvent, changeType, callbackId int,context unsafe.Pointer) {
    fmt.Printf("DeviceChangeEVENT %d\t%d\t%d\n",deviceId,changeEvent,changeType);
}

//export rawDeviceEvent
func rawDeviceEvent(data *C.char, controllerId, callbackId int,context unsafe.Pointer) {
    fmt.Printf("rawDeviceEVENT (%d):%s\n",controllerId,C.GoString(data));
}
