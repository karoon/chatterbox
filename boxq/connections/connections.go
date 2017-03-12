package connections

import (
	mgo "gopkg.in/mgo.v2"
	redis "gopkg.in/redis.v5"
)

var redisClient *redis.Client
var mgoSession *mgo.Session

// GetRedisClient returns redis client
func GetRedisClient() *redis.Client {
	if redisClient != nil {
		return redisClient
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return redisClient
}

func GetMongoSession() (*mgo.Session, error) {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.DialWithInfo(&mgo.DialInfo{
			Addrs:    []string{"localhost"},
			Database: "chatterbox",
		})
		if err != nil {
			return nil, err
		}
	}
	return mgoSession.Clone(), nil
}
