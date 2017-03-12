package mqtt

import (
	"net"

	"github.com/cihub/seelog"
)

/* Handle PUBREL */
func handlePubrel(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending DISCONNECT")
		return
	}

	clientID := (*client).Mqtt.ClientID
	clientRep := *client

	messageID := mqtt.MessageID
	seelog.Debugf("Handling PUBREL, client:(%s), messageID:(%d)", clientID, messageID)
	sendPubcomb(messageID, conn, clientRep.WriteLock)
}
