#What is ChatterBox?#
ChatterBox is an implementation of [MQTT 3.1](http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html) broker written in Go. MQTT is an excellent protocal for mobile messaging.

#Usage#
Super simple. Just run 

> go run cmd/server/main.go

The broker will start and listen on port 1883 for MQTT traffic. Command line flags:

* -p PORT: specify MQTT broker's port, default is 1883
* -r PORT: specify Redis's port, default is 6379
* -d: when set comprehensive debugging info will be printed, this may significantly harm performance.

#Dependency#
* [Redis](http://redis.io) for storage
* [Redigo](https://github.com/garyburd/redigo) as Redis binding for golang.
* [Seelog](https://github.com/cihub/seelog) for logging

#<a id="unsupported"></a>Not supported features#
* QOS level 2 is not supported. Only support QOS 1 and 0.

* Topic wildcard is not supported. Topic is always treated as plain string.
