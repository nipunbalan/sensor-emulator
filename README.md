# Sensor Emulator

This is a simple sensor emulator software implemented for IoT experiments. It emulates sensors sending data to an MQTT broker.

## Features
- This can emulate a number of sensors based on the performance of the emulation machine
### The following parameters can be defined for each sensors
- Sensor type: Ex: temperature, humidity, traffic, co2 etc.
- Frequency: The number of records per second
- Source: The source text file containing the tuples representing sensor readings

## Requirements
- Go compiler if you want to download and compile the code.
- MQTT broker for the simulator to send data to.

## Configuring the sensors
The sensors to be emulated can be configured by editing the sensoremulator.yml file
```yaml
deviceid: semu01
messaging:
  protocol: mqtt
  data_topic: sensordata
  command_topic: command1
  broker: localhost:1883
  dataratereadseconds: 10

sensors:
  sensor01:
    type: temperature
    freq: 600
    file: ./data/sensor01.dat
  sensor02:
    type: vibration
    freq: 400
    file: ./data/sensor02.dat
  sensor03:
    type: no2
    freq: 100
    file: ./data/sensor03.dat
```

## Compiling the code
```bash
git clone 
go build -o emulator main.go
```

## Running the compiled program
```bash
./emulator run
```

