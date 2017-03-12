package mqtt

import (
	"chatterbox/mqtt/auth"

	"github.com/cihub/seelog"

	"net"
	"time"
)

/* Handle PUBLISH*/
func HandlePublish(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending PUBLISH")
		return
	}

	clientID := (*client).Mqtt.ClientID
	clientRep := *client
	clientRep.UpdateLastTime()
	topic := mqtt.TopicName

	if !auth.NewUserHandler().CheckACL(clientID, topic, auth.ACLPub) {
		seelog.Debugf("client %s hasn't permission to %s on topic: %s", clientID, auth.ACLPub, topic)
		return
	}

	payload := string(mqtt.Data)
	qos := mqtt.FixedHeader.QosLevel
	retain := mqtt.FixedHeader.Retain
	messageID := mqtt.MessageID
	timestamp := time.Now().Unix()
	seelog.Debugf("Handling PUBLISH, clientID: %s, topic:(%s), payload:(%s), qos=%d, retain=%t, messageID=%d",
		clientID, topic, payload, qos, retain, messageID)

	// Create new MQTT message
	mqttMsg := CreateMqttMessage(topic, payload, clientID, qos, messageID, timestamp, retain)
	msgInternalID := mqttMsg.InternalID
	seelog.Debugf("Created new MQTT message, internal id:(%s)", msgInternalID)

	PublishMessage(mqttMsg)

	switch qos {
	case 1:
		{
			SendPuback(messageID, conn, clientRep.WriteLock)
			seelog.Debugf("PUBACK sent to client(%s)", clientID)
		}
	case 2:
		{
			SendPubrec(messageID, conn, clientRep.WriteLock)
			seelog.Debugf("PUBREC sent to client(%s)", clientID)
		}
	}
}

func PublishMessage(mqttMsg *MqttMessage) {
	topic := mqttMsg.Topic
	payload := mqttMsg.Payload
	seelog.Debugf("Publishing job, topic(%s), payload(%s)", topic, payload)
	// Update global topic record

	if mqttMsg.Retain {
		GlobalRedisClient.SetRetainMessage(topic, mqttMsg)
		seelog.Debugf("Set the message(%s) as the current retain content of topic: %s", payload, topic)
	}

	// Dispatch delivering jobs
	GlobalSubsLock.Lock()
	subs, found := GlobalSubs[topic]
	if found {
		for destID, destQos := range subs {
			go Deliver(destID, destQos, mqttMsg)
			seelog.Debugf("Started deliver job for %s", destID)
		}
	}
	GlobalSubsLock.Unlock()
	seelog.Debugf("All delivering job dispatched")
}
