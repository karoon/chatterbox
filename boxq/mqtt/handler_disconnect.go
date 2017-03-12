package mqtt

import "net"

func handleDisconnect(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending DISCONNECT")
		return
	}

	ForceDisconnect(*client, GlobalClientsLock, DONT_SEND_WILL)
}
