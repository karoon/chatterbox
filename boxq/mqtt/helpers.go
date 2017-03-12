package mqtt

import (
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cihub/seelog"
)

// This is the main place to change if we need to use channel rather than lock
func MqttSendToClient(bytes []byte, conn *net.Conn, lock *sync.Mutex) {
	if lock != nil {
		lock.Lock()
		defer func() {
			lock.Unlock()
		}()
	}
	(*conn).Write(bytes)
}

/* Checking timeout */
func CheckTimeout(client *ClientRep) {
	defer func() {
		if r := recover(); r != nil {
			seelog.Debugf("got panic, will print stack")
			debug.PrintStack()
			panic(r)
		}
	}()

	interval := client.Mqtt.KeepAliveTimer
	clientID := client.ClientID
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			now := time.Now().Unix()
			lastTimestamp := client.LastTime
			deadline := int64(float64(lastTimestamp) + float64(interval)*1.5)

			if deadline < now {
				ForceDisconnect(client, GlobalClientsLock, SEND_WILL)
				seelog.Debugf("client(%s) is timeout, kicked out",
					clientID)
			} else {
				seelog.Debugf("client(%s) will be kicked out in %d seconds",
					clientID,
					deadline-now)
			}
		case <-client.Shuttingdown:
			seelog.Debugf("client(%s) is being shutting down, stopped timeout checker", clientID)
			return
		}

	}
}

func ForceDisconnect(client *ClientRep, lock *sync.Mutex, send_will uint8) {
	if client.Disconnected == true {
		return
	}

	client.Disconnected = true

	clientID := client.Mqtt.ClientID

	seelog.Debugf("Disconnecting client(%s), clean-session:%t",
		clientID, client.Mqtt.ConnectFlags.CleanSession)

	if lock != nil {
		lock.Lock()
		seelog.Debugf("lock accuired")
	}

	delete(GlobalClients, clientID)

	if client.Mqtt.ConnectFlags.CleanSession {
		// remove her subscriptions
		seelog.Debugf("Removing subscriptions for (%s)", clientID)
		GlobalSubsLock.Lock()
		for topic := range client.Subscriptions {
			delete(GlobalSubs[topic], clientID)
			if len(GlobalSubs[topic]) == 0 {
				delete(GlobalSubs, topic)
				seelog.Debugf("last subscription of topic(%s) is removed, so this topic is removed as well", topic)
			}
		}
		showSubscriptions()
		GlobalSubsLock.Unlock()
		seelog.Debugf("Removed all subscriptions for (%s)", clientID)

		// remove her flying messages
		seelog.Debugf("Removing all flying messages for (%s)", clientID)
		GlobalRedisClient.RemoveAllFlyingMessagesForClient(clientID)
		seelog.Debugf("Removed all flying messages for (%s)", clientID)
	}

	if lock != nil {
		lock.Unlock()
		seelog.Debugf("lock released")
	}

	// FIXME: Send will if requested
	if send_will == SEND_WILL && client.Mqtt.ConnectFlags.WillFlag {
		will_topic := client.Mqtt.WillTopic
		will_payload := client.Mqtt.WillMessage
		will_qos := client.Mqtt.ConnectFlags.WillQos
		will_retain := client.Mqtt.ConnectFlags.WillRetain

		mqttMsg := CreateMqttMessage(will_topic, will_payload, clientID, will_qos,
			0, // message id won't be used here
			time.Now().Unix(), will_retain)
		publishMessage(mqttMsg)

		seelog.Debugf("Sent will for %s, topic:(%s), payload:(%s)",
			clientID, will_topic, will_payload)
	}

	client.Shuttingdown <- 1
	seelog.Debugf("Sent 1 to shutdown channel")

	seelog.Debugf("Closing socket of %s", clientID)
	(*client.Conn).Close()
}

func DeliverOnConnection(clientID string) {
	seelog.Debugf("client(%s) just reconnected, delivering on the fly messages", clientID)
	messages := GlobalRedisClient.GetFlyingMessagesForClient(clientID)
	empty := make(map[uint16]FlyingMessage)
	GlobalRedisClient.SetFlyingMessagesForClient(clientID, &empty)
	seelog.Debugf("client(%s), all flying messages put in pipeline, removed records in redis", clientID)

	for messageID, msg := range *messages {
		internalID := msg.MessageInternalID
		mqttMsg := GetMqttMessageByID(internalID)
		seelog.Debugf("re-delivering message(id=%d, internalID=%d) for %s",
			messageID, internalID, clientID)
		switch msg.Status {
		case PENDING_PUB:
			go Deliver(clientID, msg.Qos, mqttMsg)
		case PENDING_ACK:
			go Deliver(clientID, msg.Qos, mqttMsg)
		default:
			panic(fmt.Sprintf("can't re-deliver message at status(%d)", msg.Status))
		}
	}
}

// Real heavy lifting jobs for delivering message
func DeliverMessage(destClientID string, qos uint8, msg *MqttMessage) {
	GlobalClientsLock.Lock()
	clientRep, found := GlobalClients[destClientID]
	GlobalClientsLock.Unlock()

	messageID := NextOutMessageIdForClient(destClientID)
	flyMsg := CreateFlyingMessage(destClientID, msg.InternalID, qos, PENDING_PUB, messageID)

	if !found {
		GlobalRedisClient.AddFlyingMessage(destClientID, flyMsg)
		seelog.Debugf("client(%s) is offline, added flying message to Redis, message id=%d",
			destClientID, messageID)
		return
	}
	conn := clientRep.Conn
	lock := clientRep.WriteLock

	// FIXME: Add code to deal with failure
	resp := createMqtt(PUBLISH)
	resp.TopicName = msg.Topic
	if qos > 0 {
		resp.MessageID = messageID
	}
	resp.FixedHeader.QosLevel = qos
	resp.Data = []byte(msg.Payload)

	bytes, _ := encode(resp)

	lock.Lock()
	defer lock.Unlock()

	// FIXME: add write deatline
	(*conn).Write(bytes)
	seelog.Debugf("message sent by Write()")

	if qos == 1 {
		flyMsg.Status = PENDING_ACK
		GlobalRedisClient.AddFlyingMessage(destClientID, flyMsg)
		seelog.Debugf("message(msg_id=%d) sent to client(%s), waiting for ACK, added to redis",
			messageID, destClientID)
	}
}

func Deliver(destClientID string, destQos uint8, msg *MqttMessage) {
	defer func() {
		if r := recover(); r != nil {
			seelog.Debugf("got panic, will print stack")
			debug.PrintStack()
			panic(r)
		}
	}()

	seelog.Debugf("Delivering msg(internalID=%d) to client(%s)", msg.InternalID, destClientID)

	// Get effective qos: the smaller of the publisher and the subscriber
	qos := msg.Qos
	if destQos < msg.Qos {
		qos = destQos
	}

	DeliverMessage(destClientID, qos, msg)

	if qos > 0 {
		// Start retry
		go RetryDeliver(20, destClientID, qos, msg)
	}
}

func RetryDeliver(sleep uint64, destClientID string, qos uint8, msg *MqttMessage) {
	defer func() {
		if r := recover(); r != nil {
			seelog.Debugf("got panic, will print stack")
			debug.PrintStack()
			panic(r)
		}
	}()

	if sleep > 3600*4 {
		seelog.Debugf("too long retry delay(%s), abort retry deliver", sleep)
		return
	}

	time.Sleep(time.Duration(sleep) * time.Second)

	if GlobalRedisClient.IsFlyingMessagePendingAck(destClientID, msg.MessageID) {
		DeliverMessage(destClientID, qos, msg)
		seelog.Debugf("Retried delivering message %s:%d, will sleep %d seconds before next attampt",
			destClientID, msg.MessageID, sleep*2)
		RetryDeliver(sleep*2, destClientID, qos, msg)
	} else {
		seelog.Debugf("message (%s:%d) is not pending ACK, stop retry delivering",
			destClientID, msg.MessageID)
	}
}

// On connection, if clean session is set, call this method
// to clear all connections. This is the senario when previous
// CONNECT didn't set clean session bit but current one does
func RemoveAllSubscriptionsOnConnect(clientID string) {
	subs := new(map[string]uint8)
	key := fmt.Sprintf("chatterbox.client-subs.%s", clientID)
	GlobalRedisClient.Fetch(key, subs)

	GlobalRedisClient.Delete(key)

	GlobalSubsLock.Lock()
	for topic := range *subs {
		delete(GlobalSubs[topic], clientID)
	}
	GlobalSubsLock.Unlock()

}

func showSubscriptions() {
	// Disable for now
	return
	fmt.Printf("Global Subscriptions: %d topics\n", len(GlobalSubs))
	for topic, subs := range GlobalSubs {
		fmt.Printf("\t%s: %d subscriptions\n", topic, len(subs))
		for clientID, qos := range subs {
			fmt.Println("\t\t", clientID, qos)
		}
	}
}
