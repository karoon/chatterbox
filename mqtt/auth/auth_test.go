package auth

import (
	"testing"
)

var testAuthUserList = []struct {
	username string
	password string
}{
	{"unit-test-vahid", "1234567"},
	{"unit-test-saeed", "7654321"},
}

var testACLUserList = []struct {
	username string
	topic    string
	aclType  string
}{
	{"unit-test-vahid1", "groups/1", ACLPub},
	{"unit-test-vahid", "groups/2", ACLPubSub},
	{"unit-test-saeed", "groups/2", ACLPubSub},
	{"unit-test-vahid", "groups/3", ACLSub},
}

func register(t *testing.T) {
	for _, user := range testAuthUserList {
		NewUserHandler().SetUsername(user.username).SetPassword(user.password).Register()
	}
}

func TestAuthRegister(t *testing.T) {
	SetDriver(AuthDriverRedis)
	register(t)

	SetDriver(AuthDriverMongodb)
	register(t)
}

func login(t *testing.T) {
	for _, user := range testAuthUserList {
		re, err := NewUserHandler().SetUsername(user.username).SetPassword(user.password).Login()
		if err != nil {
			t.Fatalf("error %s", err)
		}
		if re == false {
			t.Fatalf("error in check authentication with %s and password %s", user.username, user.password)
		}
	}
}

func TestAuthLogin(t *testing.T) {
	SetDriver(AuthDriverRedis)
	login(t)

	SetDriver(AuthDriverMongodb)
	login(t)
}

func delete(t *testing.T) {
	return
	for _, user := range testAuthUserList {
		_, err := NewUserFromUsername(user.username)
		if err != nil {
			t.Fatalf("error in delete account %s", err)
		}
		deleted, err := NewUserHandler().SetUsername(user.username).DeleteByUsername()
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		if !deleted {
			t.Fatalf("not delete %s", user.username)
		}
	}
}

func TestAuthDelete(t *testing.T) {
	SetDriver(AuthDriverRedis)
	delete(t)

	SetDriver(AuthDriverMongodb)
	delete(t)
}

func setACL(t *testing.T) {
	for _, user := range testACLUserList {
		NewUserHandler().SetACL(user.username, user.topic, user.aclType)
	}
}

func TestACLSetACL(t *testing.T) {
	SetDriver(AuthDriverMongodb)
	setACL(t)

	SetDriver(AuthDriverRedis)
	setACL(t)
}

func checkACL(t *testing.T) {
	for _, user := range testACLUserList {
		re := NewUserHandler().CheckACL(user.username, user.topic, user.aclType)
		if re == false {
			t.Fatalf("error in check acl %s %s %s", user.username, user.topic, user.aclType)
		}
	}

	re := NewUserHandler().CheckACL("not-registered", "anonymous", ACLPub)
	if re == true {
		t.Fatalf("error in check acl for not registered user %s %s %s", "not-registered", "anonymous", ACLPub)
	}
}

func TestACLCheckACL(t *testing.T) {
	SetDriver(AuthDriverMongodb)
	checkACL(t)

	SetDriver(AuthDriverRedis)
	checkACL(t)
}

func remvoeACL(t *testing.T) {
	for _, user := range testACLUserList {
		NewUserHandler().RemoveACL(user.username, user.topic)
	}
}

func TestACLRemoveACL(t *testing.T) {
	SetDriver(AuthDriverMongodb)
	remvoeACL(t)

	SetDriver(AuthDriverRedis)
	remvoeACL(t)
}
