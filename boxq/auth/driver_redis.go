package auth

import (
	"chatterbox/boxq/connections"

	log "github.com/cihub/seelog"
	"github.com/pborman/uuid"
	redis "gopkg.in/redis.v5"
)

type RedisDriver struct {
	client *redis.Client
}

func newRedisDriver() RedisDriver {
	rd := RedisDriver{}
	rd.client = connections.GetRedisClient()
	return rd
}

func (r RedisDriver) MakeID() string {
	return uuid.New()
}

func (r RedisDriver) Register(u *User) {

	fields := make(map[string]string, 0)

	fields[authFieldID] = r.MakeID()
	fields[authFieldUsername] = u.Username
	fields[authFieldPassword] = u.PassHash()

	key := authPrefix + u.Username

	r.client.HMSet(key, fields).Result()
	return
}

func (r RedisDriver) CheckUsername(username string) (exist bool, err error) {
	key := authPrefix + username
	exist, err = r.client.Exists(key).Result()
	if err != nil {
		log.Debugf("%s", err.Error())
		return false, err
	}
	return
}

func (r RedisDriver) GetByUsername(username string) (u *User, err error) {
	key := authPrefix + username
	smap, err := r.client.HGetAll(key).Result()
	if err != nil {
		return nil, err
	}

	u = NewUserHandler()
	u.ID = smap[authFieldID]
	u.Username = username
	u.Password = smap[authFieldPassword]

	return u, nil
}

func (r RedisDriver) DeleteByUsername(username string) (deleted bool, err error) {
	key := authPrefix + username
	_, err = r.client.Del(key).Result()
	if err != nil {
		log.Debugf("%s", err.Error())
		return false, err
	}
	return true, nil
}

func (r RedisDriver) CheckACL(clientID, topic, acltype string) (actType bool, noMatch bool) {
	key := aclPrefix + clientID

	res, err := r.client.HGetAll(key).Result()
	if err != nil {
		actType = aclDeny
		return actType, false
	}

	for k, v := range res {
		if k == topic {
			if v == acltype {
				actType = aclAllow
				return actType, false
			}
		}
	}

	actType = aclDeny

	return actType, true
}

func (r RedisDriver) SetACL(clientID, topic, acltype string) {
	key := aclPrefix + clientID

	_, err := r.client.HSet(key, topic, acltype).Result()
	if err != nil {
		log.Debugf("%s", err.Error())
	}
}

func (r RedisDriver) RemoveACL(clientID, topic string) {
	key := aclPrefix + clientID
	r.client.HDel(key, topic)
}
