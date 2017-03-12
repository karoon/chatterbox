package mqtt

import (
	"chatterbox/boxq/auth"
	"fmt"
	"net"

	"github.com/cihub/seelog"
)

// Handle CONNECT
func HandleConnect(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	// mqtt.Show()
	clientID := mqtt.ClientID

	seelog.Debugf("Hanling CONNECT, client id:(%s)", clientID)

	if len(clientID) > ClientIDLimit {
		seelog.Debugf("client id(%s) is longer than %d, will send IDENTIFIER_REJECTED", clientID, ClientIDLimit)
		SendConnack(IDENTIFIER_REJECTED, conn, nil)
		return
	}

	if mqtt.ProtocolName != "MQIsdp" || mqtt.ProtocolVersion != 3 {
		seelog.Debugf("ProtocolName(%s) and/or version(%d) not supported, will send UNACCEPTABLE_PROTOCOL_VERSION",
			mqtt.ProtocolName, mqtt.ProtocolVersion)
		SendConnack(UNACCEPTABLE_PROTOCOL_VERSION, conn, nil)
		return
	}

	GlobalClientsLock.Lock()
	clientRep, existed := GlobalClients[clientID]
	if existed {
		seelog.Debugf("%s existed, will close old connection", clientID)
		ForceDisconnect(clientRep, nil, DONT_SEND_WILL)

	} else {
		seelog.Debugf("Appears to be new client, will create ClientRep")
	}

	/* Authentication */
	re, err := auth.NewUserHandler().SetUsername(mqtt.Username).SetPassword(mqtt.Password).Login()
	if err != nil {
		seelog.Debugf("error in auth")
		SendConnack(SERVER_UNAVAILABLE, conn, nil)
	}
	if re == false {
		seelog.Debugf("error in check authentication with %s and password %s", mqtt.Username, mqtt.Password)
		// auth.NewUserHandler().SetUsername(mqtt.Username).SetPassword(mqtt.Password).Register()
		SendConnack(BAD_USERNAME_OR_PASSWORD, conn, nil)
	}
	/* End of authentication */

	clientRep = CreateClientRep(clientID, conn, mqtt)

	GlobalClients[clientID] = clientRep
	GlobalClientsLock.Unlock()

	*client = clientRep
	go CheckTimeout(clientRep)
	seelog.Debugf("Timeout checker go-routine started")

	if !clientRep.Mqtt.ConnectFlags.CleanSession {
		// deliver flying messages
		DeliverOnConnection(clientID)
		// restore subscriptions to clientRep
		subs := make(map[string]uint8)
		key := fmt.Sprintf("chatterbox.client-subs.%s", clientID)
		GlobalRedisClient.Fetch(key, &subs)
		clientRep.Subscriptions = subs

	} else {
		// Remove subscriptions and flying message
		RemoveAllSubscriptionsOnConnect(clientID)
		empty := make(map[uint16]FlyingMessage)
		GlobalRedisClient.SetFlyingMessagesForClient(clientID, &empty)
	}

	SendConnack(ACCEPTED, conn, clientRep.WriteLock)
	seelog.Debugf("New client is all set and CONNACK is sent")
}
