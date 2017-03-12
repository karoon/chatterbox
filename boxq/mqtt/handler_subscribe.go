package mqtt

import (
	"chatterbox/boxq/auth"
	"fmt"
	"net"

	"github.com/cihub/seelog"
)

/* Handle SUBSCRIBE */
func HandleSubscribe(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	if *client == nil {
		panic("client_resp is nil, that means we don't have ClientRep for this client sending SUBSCRIBE")
		return
	}

	clientID := (*client).Mqtt.ClientID
	seelog.Debugf("Handling SUBSCRIBE, clientID: %s", clientID)
	clientRep := *client
	clientRep.UpdateLastTime()

	defer func() {
		GlobalSubsLock.Unlock()
		SendSuback(mqtt.MessageID, mqtt.TopicsQos, conn, clientRep.WriteLock)
	}()

	GlobalSubsLock.Lock()
	for i := 0; i < len(mqtt.Topics); i++ {

		topic := mqtt.Topics[i]
		qos := mqtt.TopicsQos[i]
		seelog.Debugf("will subscribe client(%s) to topic(%s) with qos=%d", clientID, topic, qos)

		if !auth.NewUserHandler().CheckACL(clientID, topic, auth.ACLSub) {
			seelog.Debugf("client %s hasn't permission to %s on topic: %s", clientID, "subscribe", topic)
			return
		}

		subs := GlobalSubs[topic]
		if subs == nil {
			seelog.Debugf("current subscription is the first client to topic:(%s)", topic)
			subs = make(map[string]uint8)
			GlobalSubs[topic] = subs
		}

		// FIXME: this may override existing subscription with higher QOS
		subs[clientID] = qos
		clientRep.Subscriptions[topic] = qos

		if !clientRep.Mqtt.ConnectFlags.CleanSession {
			// Store subscriptions to redis
			key := fmt.Sprintf("chatterbox.client-subs.%s", clientID)
			GlobalRedisClient.Store(key, clientRep.Subscriptions)
		}

		seelog.Debugf("finding retained message for (%s)", topic)
		retainedMsg := GlobalRedisClient.GetRetainMessage(topic)
		if retainedMsg != nil {
			go Deliver(clientID, qos, retainedMsg)
			seelog.Debugf("delivered retained message for (%s)", topic)
		}
	}

	seelog.Debugf("Subscriptions are all processed, will send SUBACK")
	showSubscriptions()
}
