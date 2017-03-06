package acl

import (
	"chatterbox/mqtt/connections"

	log "github.com/cihub/seelog"
)

const (
	aclPrefix = "mqtt_acl:"
)

const (
	aclAllow = true
	aclDeny  = false
)

const (
	AclSub    = "1" // acl for subscribe
	AclPub    = "2" // acl for publish
	AclPubSub = "3" // acl for publish subscribe
)

// CheckACL for client
func CheckACL(clientID, topic, acltype string) bool {
	rclient := connections.GetRedisClient()

	key := aclPrefix + clientID

	res, err := rclient.HGetAll(key).Result()
	if err != nil {
		return aclDeny
	}

	for k, v := range res {
		if k == topic {
			log.Debugf("topic:%s perm:%s acltype:%s", k, v, acltype)
			if v == acltype {
				log.Debugf("client:%s has permission type %s to topic:%s", clientID, acltype, topic)
				return aclAllow
			}
		}
	}

	log.Debugf("client:%s has not permission type %s to topic:%s", clientID, acltype, topic)

	return aclDeny
}

func setACL(clientID, topic, acltype string) {
	rclient := connections.GetRedisClient()

	key := aclPrefix + clientID

	b, err := rclient.HSet(key, topic, acltype).Result()
	if err != nil {
		log.Debugf("%s", err.Error())
	}

	log.Debugf("%b", b)
}

func removeACL(clientID, topic string) {
	rclient := connections.GetRedisClient()

	key := aclPrefix + clientID

	rclient.HDel(key, topic)
}
