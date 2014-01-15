#include <telldus-core.h>
#include <time.h>
#include <stdio.h>
#include <unistd.h>
#include "_cgo_export.h"

int callbackDeviceEvent = 0;
int callbackDeviceChangeEvent = 0;
int callbackRawDeviceEvent = 0;
int callbackSensorEvent = 0;

void registerCallbacks() {
    tdInit();

	callbackSensorEvent = tdRegisterSensorEvent( (TDSensorEvent)&sensorEvent, 0 );
	callbackDeviceEvent = tdRegisterDeviceEvent( (TDDeviceEvent)&deviceEvent, 0 );
	callbackDeviceChangeEvent = tdRegisterDeviceChangeEvent( (TDDeviceChangeEvent)&deviceChangeEvent, 0 );
	callbackRawDeviceEvent = tdRegisterRawDeviceEvent( (TDRawDeviceEvent)&rawDeviceEvent, 0 );
}

void unregisterCallbacks() {
	tdUnregisterCallback( callbackSensorEvent );
	tdUnregisterCallback( callbackDeviceEvent);
	tdUnregisterCallback( callbackDeviceChangeEvent );
	tdUnregisterCallback( callbackRawDeviceEvent );
	tdClose();
}
