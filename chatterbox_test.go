package chatterbox

import (
	"fmt"
	"testing"

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

func enter() {
	opts := MQTT.NewClientOptions().AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID("go-simple")
	opts.SetDefaultPublishHandler(f)

	//create and start a client using the above ClientOptions
	c = MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func benchmarkMQTT(b *testing.B, worker int) {
	if !in {
		enter()
	}

	for n := 0; n < b.N+1; n++ {
		text := fmt.Sprintf("this is msg #%d!   #%d", n, worker)
		token = c.Publish("go-mqtt/sample", 0, false, text)
	}

	defer func() {
		c.Disconnect(250)
		in = false
	}()
}

func BenchmarkMQTT1(b *testing.B) { benchmarkMQTT(b, 1) }
func BenchmarkMQTT2(b *testing.B) { benchmarkMQTT(b, 2) }
func BenchmarkMQTT3(b *testing.B) { benchmarkMQTT(b, 3) }

// func BenchmarkMQTT10(b *testing.B) { benchmarkMQTT(b, 10) }
// func BenchmarkMQTT20(b *testing.B) { benchmarkMQTT(b, 20) }
// func BenchmarkMQTT40(b *testing.B) { benchmarkMQTT(b, 40) }
