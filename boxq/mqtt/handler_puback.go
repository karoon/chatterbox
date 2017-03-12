package mqtt

import (
	"net"

	"github.com/cihub/seelog"
)

/* Handle PUBACK */
func handlePuback(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending DISCONNECT")
		return
	}

	clientID := (*client).Mqtt.ClientID
	messageID := mqtt.MessageID
	seelog.Debugf("Handling PUBACK, client:(%s), messageID:(%d)", clientID, messageID)

	messages := GlobalRedisClient.GetFlyingMessagesForClient(clientID)

	flyingMsg, found := (*messages)[messageID]

	if !found || flyingMsg.Status != PENDING_ACK {
		seelog.Debugf("message(id=%d, client=%s) is not PENDING_ACK, will ignore this PUBACK", messageID, clientID)
	} else {
		delete(*messages, messageID)
		GlobalRedisClient.SetFlyingMessagesForClient(clientID, messages)
		seelog.Debugf("acked flying message(id=%d), client:(%s)", messageID, clientID)
	}
}
