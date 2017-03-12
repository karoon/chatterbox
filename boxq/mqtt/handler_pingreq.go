package mqtt

import (
	"net"

	"github.com/cihub/seelog"
)

/* Handle PINGREQ */
func HandlePingreq(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending PINGREQ")
		return
	}

	clientID := (*client).Mqtt.ClientID
	seelog.Debugf("Handling PINGREQ, clientID: %s", clientID)
	clientRep := *client
	clientRep.UpdateLastTime()

	SendPingresp(conn, clientRep.WriteLock)
	seelog.Debugf("Sent PINGRESP, clientID: %s", clientID)
}
