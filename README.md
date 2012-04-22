gohone
======

Library for parsing process network activity on Linux.

See: https://github.com/HoneProject/Linux-Sensor

agent.go collects events emitted from the Hone kernel module and streams them at a centralized logging system using JSON.


## Building:

### Install the honeevent kernel module:
```
git clone https://github.com/HoneProject/Linux-Sensor.git honeevent
cd honeevent
make && sudo make install
sudo modprobe honeevent
```

### Go:
Make sure you have Go 1.0 installed

### Run:
`nc -l -p 7100`
`sudo go run agent.go --server localhost`
