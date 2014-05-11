#include <OneWire.h>
#include <sched.h>
#include <asm/mman.h>

// OneWire DS18S20, DS18B20, DS1822 Temperature Example
//
// http://www.pjrc.com/teensy/td_libs_OneWire.html
//
// The DallasTemperature library can do all this work for you!
// http://milesburton.com/Dallas_Temperature_Control_Library

// OneWire  ds(A0);  // on pin J12/1 (a 4.7K resistor is necessary)
OneWire  ds(GPIO3);  // on pin J11/4 (a 4.7K resistor is necessary)

// sudo modprobe pwm ; sudo modprobe sw_interrupt before running this.
void setup(void) {
    printf("Init realtime environment\n");
    if (geteuid() == 0) {
        struct sched_param sp;
        memset(&sp, 0, sizeof(sp));
        sp.sched_priority = sched_get_priority_max(SCHED_FIFO);
        sched_setscheduler(0, SCHED_FIFO, &sp);
        // mlockall(MCL_CURRENT | MCL_FUTURE);
        fprintf(stderr, "Running with realtime priority!\n");
    } else {
        fprintf(stderr, "Not running with realtime priority.\n");
    }
 
}

void loop(void) {
  byte i;
  byte present = 0;
  byte type_s;
  byte data[12];
  byte addr[8];
  float celsius, fahrenheit;
  
  if ( !ds.search(addr)) {
    printf("No more addresses.");
    printf("\n");
    ds.reset_search();
    delay(250);
    return;
  }
  
  printf("ROM =");
  for( i = 0; i < 8; i++) {
    printf(" ");
    printf("%x", addr[i]);
  }

  if (OneWire::crc8(addr, 7) != addr[7]) {
      printf("CRC is not valid!");
      return;
  }
  printf("\n");
 
  // the first ROM byte indicates which chip
  switch (addr[0]) {
    case 0x10:
      printf("  Chip = DS18S20");  // or old DS1820
      type_s = 1;
      break;
    case 0x28:
      printf("  Chip = DS18B20");
      type_s = 0;
      break;
    case 0x22:
      printf("  Chip = DS1822");
      type_s = 0;
      break;
    default:
      printf("Device is not a DS18x20 family device.");
      return;
  } 

  ds.reset();
  ds.select(addr);
  
  ds.write(0x44);
  
  // For parasitic power mode:
  // ds.write(0x44, 1);        // start conversion, with parasite power on at the end 
  // delay(1000);     // maybe 750ms is enough, maybe not
  
  present = ds.reset();
  ds.select(addr);    
  ds.write(0xBE);         // Read Scratchpad

  printf("  Data = ");
  printf("%x", present);
  printf(" ");
  for ( i = 0; i < 9; i++) {           // we need 9 bytes
    data[i] = ds.read();
    printf("%x", data[i]);
    printf(" ");
  }
  printf(" CRC=");
  printf("%x", OneWire::crc8(data, 8));
  
  if (OneWire::crc8(data, 8) != data[8]) {
    printf(" NOK");
  }
  
  printf("\n");

  // Convert the data to actual temperature
  // because the result is a 16 bit signed integer, it should
  // be stored to an "int16_t" type, which is always 16 bits
  // even when compiled on a 32 bit processor.
  int16_t raw = (data[1] << 8) | data[0];
  if (type_s) {
    raw = raw << 3; // 9 bit resolution default
    if (data[7] == 0x10) {
      // "count remain" gives full 12 bit resolution
      raw = (raw & 0xFFF0) + 12 - data[6];
    }
  } else {
    byte cfg = (data[4] & 0x60);
    // at lower res, the low bits are undefined, so let's zero them
    if (cfg == 0x00) raw = raw & ~7;  // 9 bit resolution, 93.75 ms
    else if (cfg == 0x20) raw = raw & ~3; // 10 bit res, 187.5 ms
    else if (cfg == 0x40) raw = raw & ~1; // 11 bit res, 375 ms
    // default is 12 bit resolution, 750 ms conversion time
  }
  celsius = (float)raw / 16.0;
  fahrenheit = celsius * 1.8 + 32.0;
  printf("  Temperature = ");
  printf("%f", celsius);
  printf(" Celsius, ");
  printf("%f", fahrenheit);
  printf(" Fahrenheit\n");
}
