package auth

import (
	"testing"
)

func TestMongoRegister(t *testing.T) {
	SetDriver(AuthDriverMongodb)
	for _, user := range testAuthUserList {
		NewUserHandler().SetUsername(user.username).SetPassword(user.password).Register()
	}
}

func TestMongoLogin(t *testing.T) {
	SetDriver(AuthDriverMongodb)
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

func TestMongoDelete(t *testing.T) {
	SetDriver(AuthDriverMongodb)
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
