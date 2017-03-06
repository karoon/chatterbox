package mqtt

import (
	"fmt"
	"net"

	log "github.com/cihub/seelog"
)

// Handle CONNECT
func HandleConnect(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	// mqtt.Show()
	clientID := mqtt.ClientID

	log.Debugf("Hanling CONNECT, client id:(%s)", clientID)

	if len(clientID) > ClientIDLimit {
		log.Debugf("client id(%s) is longer than %d, will send IDENTIFIER_REJECTED", clientID, ClientIDLimit)
		SendConnack(IDENTIFIER_REJECTED, conn, nil)
		return
	}

	if mqtt.ProtocolName != "MQIsdp" || mqtt.ProtocolVersion != 3 {
		log.Debugf("ProtocolName(%s) and/or version(%d) not supported, will send UNACCEPTABLE_PROTOCOL_VERSION",
			mqtt.ProtocolName, mqtt.ProtocolVersion)
		SendConnack(UNACCEPTABLE_PROTOCOL_VERSION, conn, nil)
		return
	}

	GlobalClientsLock.Lock()
	clientRep, existed := GlobalClients[clientID]
	if existed {
		log.Debugf("%s existed, will close old connection", clientID)
		ForceDisconnect(clientRep, nil, DONT_SEND_WILL)

	} else {
		log.Debugf("Appears to be new client, will create ClientRep")
	}

	clientRep = CreateClientRep(clientID, conn, mqtt)

	GlobalClients[clientID] = clientRep
	GlobalClientsLock.Unlock()

	*client = clientRep
	go CheckTimeout(clientRep)
	log.Debugf("Timeout checker go-routine started")

	if !clientRep.Mqtt.ConnectFlags.CleanSession {
		// deliver flying messages
		DeliverOnConnection(clientID)
		// restore subscriptions to clientRep
		subs := make(map[string]uint8)
		key := fmt.Sprintf("gossipd.client-subs.%s", clientID)
		GlobalRedisClient.Fetch(key, &subs)
		clientRep.Subscriptions = subs

	} else {
		// Remove subscriptions and flying message
		RemoveAllSubscriptionsOnConnect(clientID)
		empty := make(map[uint16]FlyingMessage)
		GlobalRedisClient.SetFlyingMessagesForClient(clientID, &empty)
	}

	SendConnack(ACCEPTED, conn, clientRep.WriteLock)
	log.Debugf("New client is all set and CONNACK is sent")
}