package connections

import redis "gopkg.in/redis.v5"

var redisClient *redis.Client

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
