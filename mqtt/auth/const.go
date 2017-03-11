package auth

const (
	authPrefix        = "mqtt_user:"
	authFieldPassword = "password"
	authFieldID       = "id"
	authFieldUsername = "username"
)

const (
	AuthDriverRedis int = iota + 1
	AuthDriverMongodb
)

const (
	authAllow = true
	authDeny  = false
)

const (
	aclPrefix = "mqtt_acl:"
)

const (
	aclAllow = true
	aclDeny  = false
)

const (
	// ACLSub for subscribe id
	ACLSub = "1"
	// ACLPub for publish id
	ACLPub = "2"
	// ACLPubSub for publish id
	ACLPubSub = "3"
)

const (
	ACLDriverRedis int = iota + 1
	ACLDriverMongodb
)
