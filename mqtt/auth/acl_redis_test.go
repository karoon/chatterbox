package auth

import (
	"testing"
)

func TestACLRedisSetACL(t *testing.T) {
	SetDriver(AuthDriverRedis)
	for _, user := range testACLUserList {
		NewUserHandler().SetACL(user.username, user.topic, user.aclType)
	}
}

func TestACLRedisCheckACL(t *testing.T) {
	SetDriver(AuthDriverRedis)
	for _, user := range testACLUserList {
		re := NewUserHandler().CheckACL(user.username, user.topic, user.aclType)
		if re == false {
			t.Fatalf("error in check acl %s %s %s", user.username, user.topic, user.aclType)
		}
	}
}

// func TestACLRedisRemoveACL(t *testing.T) {
// 	SetDriver(AuthDriverRedis)
// 	for _, user := range testACLUserList {
// 		NewUserHandler().RemoveACL(user.username, user.topic)
// 	}
// }
