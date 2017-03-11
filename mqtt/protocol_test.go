package mqtt

import (
	"fmt"
	"os"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func getConnection() MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID("unit-test-saeed")
	opts.SetUsername("unit-test-saeed")
	opts.SetPassword("7654321")
	opts.SetDefaultPublishHandler(f)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	return c

}

func TestSub(t *testing.T) {
	c := getConnection()
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("%s", token.Error())
	}

	//subscribe to the topic /go-mqtt/sample and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription
	if token := c.Subscribe("groups/2", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	//unsubscribe from /go-mqtt/sample
	if token := c.Unsubscribe("groups/2"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	c.Disconnect(250)

}

func TestPublish(t *testing.T) {
	c := getConnection()
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("%s", token.Error())
	}

	//Publish 5 messages to /go-mqtt/sample at qos 1 and wait for the receipt
	//from the server after sending each message
	for i := 0; i < 5; i++ {
		text := fmt.Sprintf("this is msg #%d!", i)
		token := c.Publish("groups/2", 0, false, text)
		if token.Error() != nil {
			t.Logf("%s", token.Error())
		}

		token.Wait()
	}

	time.Sleep(3 * time.Second)

	c.Disconnect(250)
}
