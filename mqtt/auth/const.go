package auth

const (
	authPrefix        = "mqtt_user:"
	authFieldPassword = "password"
	authFieldID       = "id"
	authFieldUsername = "username"
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
	AclSub    = "1" // acl for subscribe
	AclPub    = "2" // acl for publish
	AclPubSub = "3" // acl for publish subscribe
)

const (
	ACLDriverRedis int = iota + 1
	ACLDriverMongodb
)
