package mqtt

import (
	"net"
	"sync"
)

func sendConnack(rc uint8, conn *net.Conn, lock *sync.Mutex) {
	resp := createMqtt(CONNACK)
	resp.ReturnCode = rc

	bytes, _ := encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

func sendPubrec(messageID uint16, conn *net.Conn, lock *sync.Mutex) {
	resp := createMqtt(PUBREC)
	resp.MessageID = messageID
	bytes, _ := encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

func sendPuback(msgID uint16, conn *net.Conn, lock *sync.Mutex) {
	resp := createMqtt(PUBACK)
	resp.MessageID = msgID
	bytes, _ := encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

func sendSuback(msgID uint16, qos_list []uint8, conn *net.Conn, lock *sync.Mutex) {
	resp := createMqtt(SUBACK)
	resp.MessageID = msgID
	resp.TopicsQos = qos_list

	bytes, _ := encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

func sendUnsuback(msgID uint16, conn *net.Conn, lock *sync.Mutex) {
	resp := createMqtt(UNSUBACK)
	resp.MessageID = msgID
	bytes, _ := encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

func sendPingresp(conn *net.Conn, lock *sync.Mutex) {
	resp := createMqtt(PINGRESP)
	bytes, _ := encode(resp)
	MqttSendToClient(bytes, conn, lock)
}

func sendPubcomb(msgID uint16, conn *net.Conn, lock *sync.Mutex) {
	resp := createMqtt(PUBCOMP)
	resp.MessageID = msgID
	bytes, _ := encode(resp)
	MqttSendToClient(bytes, conn, lock)
}
