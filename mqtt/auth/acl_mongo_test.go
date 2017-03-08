package auth

import (
	"testing"
)

func TestACLMongoSetACL(t *testing.T) {
	SetDriver(AuthDriverMongodb)
	for _, user := range testACLUserList {
		NewUserHandler().SetACL(user.username, user.topic, user.aclType)
	}
}

func TestACLMongoCheckACL(t *testing.T) {
	SetDriver(AuthDriverMongodb)
	for _, user := range testACLUserList {
		re := NewUserHandler().CheckACL(user.username, user.topic, user.aclType)
		if re == false {
			t.Fatalf("error in check acl %s %s %s", user.username, user.topic, user.aclType)
		}
	}
}

func TestACLMongoRemoveACL(t *testing.T) {
	SetDriver(AuthDriverMongodb)
	for _, user := range testACLUserList {
		NewUserHandler().RemoveACL(user.username, user.topic)
	}
}
