package main

import (
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	// fmt.Printf("TOPIC: %s\n", msg.Topic())
	// fmt.Printf("MSG: %s\n", msg.Payload())
}

var c MQTT.Client
var in bool = false
var token MQTT.Token

func init() {
	opts := MQTT.NewClientOptions().AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID("go-simple")
	opts.SetDefaultPublishHandler(f)

	c = MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func main() {
	c.Publish("go-mqtt/sample", 0, false, "-------------------------------------------")
	for n := 0; n < 1; n++ {
		text := fmt.Sprintf("this is msg #%d!   #%d", n, 1)
		token = c.Publish("go-mqtt/sample", 1, false, text)
	}

	time.Sleep(1 * time.Second)

	defer c.Disconnect(250)
}
