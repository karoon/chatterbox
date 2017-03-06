package mqtt

import redis "gopkg.in/redis.v5"

var redisClient *redis.Client

func getRedisClient() *redis.Client {
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
