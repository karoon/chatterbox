package mqtt

import (
	"net"

	log "github.com/cihub/seelog"
)

/* Handle PUBREL */
func HandlePubrel(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending DISCONNECT")
		return
	}

	clientID := (*client).Mqtt.ClientID
	clientRep := *client

	messageID := mqtt.MessageID
	log.Debugf("Handling PUBREL, client:(%s), messageID:(%d)", clientID, messageID)
	SendPubcomb(messageID, conn, clientRep.WriteLock)
}
