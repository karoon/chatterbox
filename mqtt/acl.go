package mqtt

import log "github.com/cihub/seelog"

const (
	aclPrefix = "mqtt_acl:"
)

const (
	aclSub    = "1" // acl for subscribe
	aclPub    = "2" // acl for publish
	aclPubSub = "3" // acl for publish subscribe
)

func checkACL(clientID string, topic string, acltype string) bool {
	rclient := getRedisClient()

	key := aclPrefix + clientID

	res, err := rclient.HGetAll(key).Result()
	if err != nil {
		return false
	}

	for k, v := range res {
		if k == topic {
			log.Debugf("topic:%s perm:%s acltype:%s", k, v, acltype)
			if v == acltype {
				log.Debugf("client:%s has permission type %s to topic:%s", clientID, acltype, topic)
				return true
			}
		}
	}

	log.Debugf("client:%s has not permission type %s to topic:%s", clientID, acltype, topic)

	return false
}

func setACL(clientID string, topic string, acltype string) {
	rclient := getRedisClient()

	key := aclPrefix + clientID

	b, err := rclient.HSet(key, topic, acltype).Result()
	if err != nil {
		log.Debugf("%s", err.Error())
	}

	log.Debugf("%b", b)
}

func removeACL(clientID string, topic string) {
	rclient := getRedisClient()

	key := aclPrefix + clientID

	rclient.HDel(key, topic)
}
