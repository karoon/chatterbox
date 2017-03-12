package auth

import "chatterbox/mqtt/types"

// import "mqtt/types"

const (
	authPrefix        = "mqtt_user:"
	authFieldPassword = "password"
	authFieldID       = "id"
	authFieldUsername = "username"
)

const (
	AuthDriverRedis   types.AuthDriverType = types.AuthDriverTypeRedis
	AuthDriverMongodb types.AuthDriverType = types.AuthDriverTypeMongodb
)

const (
	authAllow types.AuthActionType = types.AuthActionTypeAllow
	authDeny  types.AuthActionType = types.AuthActionTypeDeny
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
