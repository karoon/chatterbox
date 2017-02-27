package mqtt

import (
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/cihub/seelog"
)

const (
	SEND_WILL = uint8(iota)
	DONT_SEND_WILL
)

// Handle CONNECT
func HandleConnect(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	//mqtt.Show()
	clientID := mqtt.ClientID

	log.Debugf("Hanling CONNECT, client id:(%s)", clientID)

	if len(clientID) > 23 {
		log.Debugf("client id(%s) is longer than 23, will send IDENTIFIER_REJECTED", clientID)
		SendConnack(IDENTIFIER_REJECTED, conn, nil)
		return
	}

	if mqtt.ProtocolName != "MQIsdp" || mqtt.ProtocolVersion != 3 {
		log.Debugf("ProtocolName(%s) and/or version(%d) not supported, will send UNACCEPTABLE_PROTOCOL_VERSION",
			mqtt.ProtocolName, mqtt.ProtocolVersion)
		SendConnack(UNACCEPTABLE_PROTOCOL_VERSION, conn, nil)
		return
	}

	G_clients_lock.Lock()
	clientRep, existed := G_clients[clientID]
	if existed {
		log.Debugf("%s existed, will close old connection", clientID)
		ForceDisconnect(clientRep, nil, DONT_SEND_WILL)

	} else {
		log.Debugf("Appears to be new client, will create ClientRep")
	}

	clientRep = CreateClientRep(clientID, conn, mqtt)

	G_clients[clientID] = clientRep
	G_clients_lock.Unlock()

	*client = clientRep
	go CheckTimeout(clientRep)
	log.Debugf("Timeout checker go-routine started")

	if !clientRep.Mqtt.ConnectFlags.CleanSession {
		// deliver flying messages
		DeliverOnConnection(clientID)
		// restore subscriptions to clientRep
		subs := make(map[string]uint8)
		key := fmt.Sprintf("gossipd.client-subs.%s", clientID)
		G_redis_client.Fetch(key, &subs)
		clientRep.Subscriptions = subs

	} else {
		// Remove subscriptions and flying message
		RemoveAllSubscriptionsOnConnect(clientID)
		empty := make(map[uint16]FlyingMessage)
		G_redis_client.SetFlyingMessagesForClient(clientID, &empty)
	}

	SendConnack(ACCEPTED, conn, clientRep.WriteLock)
	log.Debugf("New client is all set and CONNACK is sent")
}

func SendConnack(rc uint8, conn *net.Conn, lock *sync.Mutex) {
	resp := CreateMqtt(CONNACK)
	resp.ReturnCode = rc

	bytes, _ := Encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

/* Handle PUBLISH*/
// FIXME: support qos = 2
func HandlePublish(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending PUBLISH")
		return
	}

	clientID := (*client).Mqtt.ClientID
	clientRep := *client
	clientRep.UpdateLastTime()
	topic := mqtt.TopicName
	payload := string(mqtt.Data)
	qos := mqtt.FixedHeader.QosLevel
	retain := mqtt.FixedHeader.Retain
	messageID := mqtt.MessageID
	timestamp := time.Now().Unix()
	log.Debugf("Handling PUBLISH, clientID: %s, topic:(%s), payload:(%s), qos=%d, retain=%t, messageID=%d",
		clientID, topic, payload, qos, retain, messageID)

	// Create new MQTT message
	mqttMsg := CreateMqttMessage(topic, payload, clientID, qos, messageID, timestamp, retain)
	msgInternalID := mqttMsg.InternalID
	log.Debugf("Created new MQTT message, internal id:(%s)", msgInternalID)

	PublishMessage(mqttMsg)

	// Send PUBACK if QOS is 1
	if qos == 1 {
		SendPuback(messageID, conn, clientRep.WriteLock)
		log.Debugf("PUBACK sent to client(%s)", clientID)
	}
}

func SendPuback(msg_id uint16, conn *net.Conn, lock *sync.Mutex) {
	resp := CreateMqtt(PUBACK)
	resp.MessageID = msg_id
	bytes, _ := Encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

/* Handle SUBSCRIBE */
func HandleSubscribe(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending SUBSCRIBE")
		return
	}

	clientID := (*client).Mqtt.ClientID
	log.Debugf("Handling SUBSCRIBE, clientID: %s", clientID)
	clientRep := *client
	clientRep.UpdateLastTime()

	defer func() {
		G_subs_lock.Unlock()
		SendSuback(mqtt.MessageID, mqtt.TopicsQos, conn, clientRep.WriteLock)
	}()

	G_subs_lock.Lock()
	for i := 0; i < len(mqtt.Topics); i++ {
		topic := mqtt.Topics[i]
		qos := mqtt.TopicsQos[i]
		log.Debugf("will subscribe client(%s) to topic(%s) with qos=%d",
			clientID, topic, qos)

		subs := G_subs[topic]
		if subs == nil {
			log.Debugf("current subscription is the first client to topic:(%s)", topic)
			subs = make(map[string]uint8)
			G_subs[topic] = subs
		}

		// FIXME: this may override existing subscription with higher QOS
		subs[clientID] = qos
		clientRep.Subscriptions[topic] = qos

		if !clientRep.Mqtt.ConnectFlags.CleanSession {
			// Store subscriptions to redis
			key := fmt.Sprintf("gossipd.client-subs.%s", clientID)
			G_redis_client.Store(key, clientRep.Subscriptions)
		}

		log.Debugf("finding retained message for (%s)", topic)
		retainedMsg := G_redis_client.GetRetainMessage(topic)
		if retainedMsg != nil {
			go Deliver(clientID, qos, retainedMsg)
			log.Debugf("delivered retained message for (%s)", topic)
		}
	}
	log.Debugf("Subscriptions are all processed, will send SUBACK")
	showSubscriptions()
}

func SendSuback(msg_id uint16, qos_list []uint8, conn *net.Conn, lock *sync.Mutex) {
	resp := CreateMqtt(SUBACK)
	resp.MessageID = msg_id
	resp.TopicsQos = qos_list

	bytes, _ := Encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

/* Handle UNSUBSCRIBE */
func HandleUnsubscribe(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending UNSUBSCRIBE")
		return
	}

	clientID := (*client).Mqtt.ClientID
	log.Debugf("Handling UNSUBSCRIBE, clientID: %s", clientID)
	clientRep := *client
	clientRep.UpdateLastTime()

	defer func() {
		G_subs_lock.Unlock()
		SendUnsuback(mqtt.MessageID, conn, clientRep.WriteLock)
	}()

	G_subs_lock.Lock()
	for i := 0; i < len(mqtt.Topics); i++ {
		topic := mqtt.Topics[i]

		log.Debugf("unsubscribing client(%s) from topic(%s)",
			clientID, topic)

		delete(clientRep.Subscriptions, topic)

		subs := G_subs[topic]
		if subs == nil {
			log.Debugf("topic(%s) has no subscription, no need to unsubscribe", topic)
		} else {
			delete(subs, clientID)
			if len(subs) == 0 {
				delete(G_subs, topic)
				log.Debugf("last subscription of topic(%s) is removed, so this topic is removed as well", topic)
			}
		}
	}
	log.Debugf("unsubscriptions are all processed, will send UNSUBACK")

	showSubscriptions()
}

func SendUnsuback(msg_id uint16, conn *net.Conn, lock *sync.Mutex) {
	resp := CreateMqtt(UNSUBACK)
	resp.MessageID = msg_id
	bytes, _ := Encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

/* Handle PINGREQ */

func HandlePingreq(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending PINGREQ")
		return
	}

	clientID := (*client).Mqtt.ClientID
	log.Debugf("Handling PINGREQ, clientID: %s", clientID)
	clientRep := *client
	clientRep.UpdateLastTime()

	SendPingresp(conn, clientRep.WriteLock)
	log.Debugf("Sent PINGRESP, clientID: %s", clientID)
}

func SendPingresp(conn *net.Conn, lock *sync.Mutex) {
	resp := CreateMqtt(PINGRESP)
	bytes, _ := Encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

/* Handle DISCONNECT */

func HandleDisconnect(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending DISCONNECT")
		return
	}

	ForceDisconnect(*client, G_clients_lock, DONT_SEND_WILL)
}

/* Handle PUBACK */
func HandlePuback(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending DISCONNECT")
		return
	}

	clientID := (*client).Mqtt.ClientID
	messageID := mqtt.MessageID
	log.Debugf("Handling PUBACK, client:(%s), messageID:(%d)", clientID, messageID)

	messages := G_redis_client.GetFlyingMessagesForClient(clientID)

	flying_msg, found := (*messages)[messageID]

	if !found || flying_msg.Status != PENDING_ACK {
		log.Debugf("message(id=%d, client=%s) is not PENDING_ACK, will ignore this PUBACK",
			messageID, clientID)
	} else {
		delete(*messages, messageID)
		G_redis_client.SetFlyingMessagesForClient(clientID, messages)
		log.Debugf("acked flying message(id=%d), client:(%s)", messageID, clientID)
	}
}

/* Helper functions */

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
			log.Debugf("got panic, will print stack")
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
				ForceDisconnect(client, G_clients_lock, SEND_WILL)
				log.Debugf("client(%s) is timeout, kicked out",
					clientID)
			} else {
				log.Debugf("client(%s) will be kicked out in %d seconds",
					clientID,
					deadline-now)
			}
		case <-client.Shuttingdown:
			log.Debugf("client(%s) is being shutting down, stopped timeout checker", clientID)
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

	log.Debugf("Disconnecting client(%s), clean-session:%t",
		clientID, client.Mqtt.ConnectFlags.CleanSession)

	if lock != nil {
		lock.Lock()
		log.Debugf("lock accuired")
	}

	delete(G_clients, clientID)

	if client.Mqtt.ConnectFlags.CleanSession {
		// remove her subscriptions
		log.Debugf("Removing subscriptions for (%s)", clientID)
		G_subs_lock.Lock()
		for topic := range client.Subscriptions {
			delete(G_subs[topic], clientID)
			if len(G_subs[topic]) == 0 {
				delete(G_subs, topic)
				log.Debugf("last subscription of topic(%s) is removed, so this topic is removed as well", topic)
			}
		}
		showSubscriptions()
		G_subs_lock.Unlock()
		log.Debugf("Removed all subscriptions for (%s)", clientID)

		// remove her flying messages
		log.Debugf("Removing all flying messages for (%s)", clientID)
		G_redis_client.RemoveAllFlyingMessagesForClient(clientID)
		log.Debugf("Removed all flying messages for (%s)", clientID)
	}

	if lock != nil {
		lock.Unlock()
		log.Debugf("lock released")
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
		PublishMessage(mqttMsg)

		log.Debugf("Sent will for %s, topic:(%s), payload:(%s)",
			clientID, will_topic, will_payload)
	}

	client.Shuttingdown <- 1
	log.Debugf("Sent 1 to shutdown channel")

	log.Debugf("Closing socket of %s", clientID)
	(*client.Conn).Close()
}

func PublishMessage(mqttMsg *MqttMessage) {
	topic := mqttMsg.Topic
	payload := mqttMsg.Payload
	log.Debugf("Publishing job, topic(%s), payload(%s)", topic, payload)
	// Update global topic record

	if mqttMsg.Retain {
		G_redis_client.SetRetainMessage(topic, mqttMsg)
		log.Debugf("Set the message(%s) as the current retain content of topic:%s", payload, topic)
	}

	// Dispatch delivering jobs
	G_subs_lock.Lock()
	subs, found := G_subs[topic]
	if found {
		for dest_id, dest_qos := range subs {
			go Deliver(dest_id, dest_qos, mqttMsg)
			log.Debugf("Started deliver job for %s", dest_id)
		}
	}
	G_subs_lock.Unlock()
	log.Debugf("All delivering job dispatched")
}

func DeliverOnConnection(clientID string) {
	log.Debugf("client(%s) just reconnected, delivering on the fly messages", clientID)
	messages := G_redis_client.GetFlyingMessagesForClient(clientID)
	empty := make(map[uint16]FlyingMessage)
	G_redis_client.SetFlyingMessagesForClient(clientID, &empty)
	log.Debugf("client(%s), all flying messages put in pipeline, removed records in redis", clientID)

	for messageID, msg := range *messages {
		internalID := msg.MessageInternalID
		mqttMsg := GetMqttMessageByID(internalID)
		log.Debugf("re-delivering message(id=%d, internalID=%d) for %s",
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
func DeliverMessage(dest_clientID string, qos uint8, msg *MqttMessage) {
	G_clients_lock.Lock()
	clientRep, found := G_clients[dest_clientID]
	G_clients_lock.Unlock()
	var conn *net.Conn
	var lock *sync.Mutex
	messageID := NextOutMessageIdForClient(dest_clientID)
	flyMsg := CreateFlyingMessage(dest_clientID, msg.InternalID, qos, PENDING_PUB, messageID)

	if found {
		conn = clientRep.Conn
		lock = clientRep.WriteLock
	} else {
		G_redis_client.AddFlyingMessage(dest_clientID, flyMsg)
		log.Debugf("client(%s) is offline, added flying message to Redis, message id=%d",
			dest_clientID, messageID)
		return
	}

	// FIXME: Add code to deal with failure
	resp := CreateMqtt(PUBLISH)
	resp.TopicName = msg.Topic
	if qos > 0 {
		resp.MessageID = messageID
	}
	resp.FixedHeader.QosLevel = qos
	resp.Data = []byte(msg.Payload)

	bytes, _ := Encode(resp)

	lock.Lock()
	defer func() {
		lock.Unlock()
	}()
	// FIXME: add write deatline
	(*conn).Write(bytes)
	log.Debugf("message sent by Write()")

	if qos == 1 {
		flyMsg.Status = PENDING_ACK
		G_redis_client.AddFlyingMessage(dest_clientID, flyMsg)
		log.Debugf("message(msg_id=%d) sent to client(%s), waiting for ACK, added to redis",
			messageID, dest_clientID)
	}
}

func Deliver(dest_clientID string, dest_qos uint8, msg *MqttMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Debugf("got panic, will print stack")
			debug.PrintStack()
			panic(r)
		}
	}()

	log.Debugf("Delivering msg(internalID=%d) to client(%s)", msg.InternalID, dest_clientID)

	// Get effective qos: the smaller of the publisher and the subscriber
	qos := msg.Qos
	if dest_qos < msg.Qos {
		qos = dest_qos
	}

	DeliverMessage(dest_clientID, qos, msg)

	if qos > 0 {
		// Start retry
		go RetryDeliver(20, dest_clientID, qos, msg)
	}
}

func RetryDeliver(sleep uint64, dest_clientID string, qos uint8, msg *MqttMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Debugf("got panic, will print stack")
			debug.PrintStack()
			panic(r)
		}
	}()

	if sleep > 3600*4 {
		log.Debugf("too long retry delay(%s), abort retry deliver", sleep)
		return
	}

	time.Sleep(time.Duration(sleep) * time.Second)

	if G_redis_client.IsFlyingMessagePendingAck(dest_clientID, msg.MessageID) {
		DeliverMessage(dest_clientID, qos, msg)
		log.Debugf("Retried delivering message %s:%d, will sleep %d seconds before next attampt",
			dest_clientID, msg.MessageID, sleep*2)
		RetryDeliver(sleep*2, dest_clientID, qos, msg)
	} else {
		log.Debugf("message (%s:%d) is not pending ACK, stop retry delivering",
			dest_clientID, msg.MessageID)
	}
}

// On connection, if clean session is set, call this method
// to clear all connections. This is the senario when previous
// CONNECT didn't set clean session bit but current one does
func RemoveAllSubscriptionsOnConnect(clientID string) {
	subs := new(map[string]uint8)
	key := fmt.Sprintf("gossipd.client-subs.%s", clientID)
	G_redis_client.Fetch(key, subs)

	G_redis_client.Delete(key)

	G_subs_lock.Lock()
	for topic := range *subs {
		delete(G_subs[topic], clientID)
	}
	G_subs_lock.Unlock()

}

func showSubscriptions() {
	// Disable for now
	return
	fmt.Printf("Global Subscriptions: %d topics\n", len(G_subs))
	for topic, subs := range G_subs {
		fmt.Printf("\t%s: %d subscriptions\n", topic, len(subs))
		for clientID, qos := range subs {
			fmt.Println("\t\t", clientID, qos)
		}
	}
}
