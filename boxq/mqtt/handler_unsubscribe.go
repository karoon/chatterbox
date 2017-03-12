package mqtt

import (
	"net"

	"github.com/cihub/seelog"
)

/* Handle UNSUBSCRIBE */
func HandleUnsubscribe(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending UNSUBSCRIBE")
		return
	}

	clientID := (*client).Mqtt.ClientID
	seelog.Debugf("Handling UNSUBSCRIBE, clientID: %s", clientID)
	clientRep := *client
	clientRep.UpdateLastTime()

	defer func() {
		GlobalSubsLock.Unlock()
		SendUnsuback(mqtt.MessageID, conn, clientRep.WriteLock)
	}()

	GlobalSubsLock.Lock()
	for i := 0; i < len(mqtt.Topics); i++ {
		topic := mqtt.Topics[i]

		seelog.Debugf("unsubscribing client(%s) from topic(%s)",
			clientID, topic)

		delete(clientRep.Subscriptions, topic)

		subs := GlobalSubs[topic]
		if subs == nil {
			seelog.Debugf("topic(%s) has no subscription, no need to unsubscribe", topic)
		} else {
			delete(subs, clientID)
			if len(subs) == 0 {
				delete(GlobalSubs, topic)
				seelog.Debugf("last subscription of topic(%s) is removed, so this topic is removed as well", topic)
			}
		}
	}
	seelog.Debugf("unsubscriptions are all processed, will send UNSUBACK")

	showSubscriptions()
}
