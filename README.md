# OneWire Library
This is just port of the [OneWire v2.2 Library](http://www.pjrc.com/teensy/td_libs_OneWire.html) for the [pcDuino3](http://www.pcduino.com/pcduino-v3/)

After I bought the pcDuino3, I've realized there's no library available for the OneWire protocol so I decided to port it myself.

# Observations

I have tested this with 4 DS18B20 sensors (waterproof with about 1 meter of cable each) and they work, but there are many cases that I haven't tested because I didn't have a need for them:
1. Parasitic power mode
2. Other OneWire devices (I don't have others)
3. Accessing the sensors directly based on a known ROM ID isn't tested (plenty of space on pcDuino than to do micromanagement)

Also, code was changed more than what was required to get it to compile; the chages were needed because of differences between the A20 CPU and Atmel:

1. On the library itself the bus was taken low before its direction was changed. This doesn't work on pcDuino, you have to change the pin value after you set its direction.
2. delayMicrosecond had to be reimplemented and I've used a version from here: http://www.raspberry-projects.com/pi/programming-in-c/timing/clock_gettime-for-acurate-timing. For more explanations why this is needed, please check this thread: http://pcduino.com/forum/index.php?topic=4438.msg6678 or the Timings section below.
3. Skip using pinMode and rely on hw_pinMode if possible. If you check the c_environment, you will see that pinMode implementation actually tries to disable the a possible running PWM timer. This slows down the execution and breaks the OneWire protocol timings. In case you need a PWM pin shared with the OneWire protocol, try to disable the PWM ahead of the OneWire communications and then rely on hw_pinMode to do the direction change.

Other changes are mentioned within the *OneWire.cpp* file.

You might get errors that /proc/adc0 is unreadable. Please use:

    sudo modprobe adc
    
to get around this.

Also, during the write() operations (in normal Power mode, not Parasitic) the bus was brought LOW after the having its direction changed as INPUT. If lenghtly operations come after this, then this would reset the sensors - unwanted scenario in my case.

# Timings

I had a difficult time getting the timings right for the OneWire protocol. Using the realtime clocks in the kernel, I tried to time how much it takes to do 10 digitalWrite to a pin. Time varies between 2 and 3uS (microseconds), so these delays have to also be taken into account when applying the protocol delays specified here: http://www.maximintegrated.com/app-notes/index.mvp/id/126

Current implementation of delayMicroseconds on pcDuino's c_environment relies on the usleep call. On a OS with real time processes going on, multi threaded, this only guarantees the minimum sleep time, but dues not guarentee that the thread is given back the control immediatly after the sleep expires. Hopefully future versions of the NAND image for pcDuino will come with a fined tuned busy loop for accurate delays.

From my experiments, about 2-3 bus communications out of 10 will end up with a bad CRC (I suppose due to the fact that you can't perfectly control the busy wait timings when other RT process come into play). It might work better using the timers provided by A20 chip on pcDuino3, however, I haven't investigated more.

You will see that the provided *examples/DS18x20_Temperature.c* tries to rise the priority of the process to RT (Real Time) and you need to execute it as root (sudo). Without the RT priority, there were visibily more CRC failures and I had to resort to multiple reads just to reduce the risk of having CRC on the single read.

With more robust architectures (where many more sensors are present and lower CRC errors would minimize the measurement time) I would go with a dedicated OneWire master.

# Pull-up Resistor

I resorted to use a 10K potentiometer set at around 5K mark so I can have a way to adjusts for my sensors. However it seems it didn't matter that much if it was 2K or 7k from few trial and errors. Lower or higher and you start getting invalid results.

# Goals

I intend to use pcDuino as a data logger for my 4 aquariums since I had some bad heaters in the past. I wanted a smarter solution than to stack a lot of Arduino shields just to get some Wireless going.

Right now, two sensors are connected to pcDuino which answers queries over SSH from the Cacti server running on the NAS. pcDuino is defined as a separate host (I haven't got around activating SNMP on it yet) and temperatures are plotted on a graph. I will have a 5th sensor to have the room temperature and see the link between the room temperature and the aquariums temperatures.

Also, pcDuino will display temperatures itself using the [Arduino 1.77' TFT LCD](http://arduino.cc/en/Main/GTFT) - I have ported that library too, you can find the source code here: https://bitbucket.org/viulian/arduino-tft-library-for-pcduino3 Even if you have graphs on a web site, it is much more easier to get a picture that everything looks alright if you can catch a glimpse of the temperature readings as you walk by the tanks :)

I will probably publish the software (including Cacti templates) after I get the whole thing assembled more nicely.
