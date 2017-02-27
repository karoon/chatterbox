package mqtt

import (
	"fmt"
	"sync"
	"sync/atomic"
)

/*
 This is the type represents a message received from publisher.
 FlyingMessage(message should be delivered to specific subscribers)
 reference MqttMessage
*/
type MqttMessage struct {
	Topic          string
	Payload        string
	Qos            uint8
	SenderClientID string
	MessageID      uint16
	InternalID     uint64
	CreatedAt      int64
	Retain         bool
}

func (msg *MqttMessage) Show() {
	fmt.Printf("MQTT Message:\n")
	fmt.Println("Topic:", msg.Topic)
	fmt.Println("Payload:", msg.Payload)
	fmt.Println("Qos:", msg.Qos)
	fmt.Println("SenderClientID:", msg.SenderClientID)
	fmt.Println("MessageID:", msg.MessageID)
	fmt.Println("InternalID:", msg.InternalID)
	fmt.Println("CreatedAt:", msg.CreatedAt)
	fmt.Println("Retain:", msg.Retain)
}

func (msg *MqttMessage) RedisKey() string {
	return fmt.Sprintf("gossipd.mqtt-msg.%d", msg.InternalID)
}

func (msg *MqttMessage) Store() {
	key := msg.RedisKey()
	GlobalRedisClient.Store(key, msg)
	GlobalRedisClient.Expire(key, 7*24*3600)
}

// InternalID -> Message
// FIXME: Add code to store G_messages to disk
var G_messages map[uint64]*MqttMessage = make(map[uint64]*MqttMessage)
var G_messages_lock *sync.Mutex = new(sync.Mutex)

func CreateMqttMessage(topic, payload, sender_id string,
	qos uint8, message_id uint16,
	created_at int64, retain bool) *MqttMessage {

	msg := new(MqttMessage)
	msg.Topic = topic
	msg.Payload = payload
	msg.Qos = qos
	msg.SenderClientID = sender_id
	msg.MessageID = message_id
	msg.InternalID = GetNextMessageInternalID()
	msg.CreatedAt = created_at
	msg.Retain = retain

	G_messages_lock.Lock()
	G_messages[msg.InternalID] = msg
	G_messages_lock.Unlock()

	msg.Store()

	return msg
}

var gNextMqttMessageInternalID uint64

func GetNextMessageInternalID() uint64 {
	return atomic.AddUint64(&gNextMqttMessageInternalID, 1)
}

// This is thread-safe
func GetMqttMessageByID(internalID uint64) *MqttMessage {
	key := fmt.Sprintf("gossipd.mqtt-msg.%d", internalID)

	msg := new(MqttMessage)
	GlobalRedisClient.Fetch(key, msg)
	return msg
}

/*
 This is the type represents a message should be delivered to
 specific client
*/
type FlyingMessage struct {
	Qos               uint8 // the Qos in effect
	DestClientID      string
	MessageInternalID uint64 // The MqttMessage of interest
	Status            uint8  // The status of this message, like PENDING_PUB(deliver occured
	// when client if offline), PENDING_ACK, etc
	ClientMessageID uint16 // The message id to be used in MQTT packet
}

const (
	PENDING_PUB = uint8(iota + 1)
	PENDING_ACK
)

func CreateFlyingMessage(dest_id string, messageInternalID uint64, qos, status uint8, message_id uint16) *FlyingMessage {
	msg := new(FlyingMessage)
	msg.Qos = qos
	msg.DestClientID = dest_id
	msg.MessageInternalID = messageInternalID
	msg.Status = status
	msg.ClientMessageID = message_id
	return msg
}
