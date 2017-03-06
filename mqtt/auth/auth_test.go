package auth

import (
	"testing"
)

var userList = []struct {
	username string
	password string
}{
	{"unit-test-vahid", "1234567"},
	{"unit-test-saeed", "7654321"},
}

func TestRegister(t *testing.T) {
	for _, user := range userList {
		NewUserHandler().SetUsername(user.username).SetPassword(user.password).Register()
	}
}

func TestLogin(t *testing.T) {
	for _, user := range userList {
		re, err := NewUserHandler().SetUsername(user.username).SetPassword(user.password).Login()
		if err != nil {
			t.Fatalf("error %s", err)
		}
		if re == false {
			t.Fatalf("error in check authentication with %s and password %s", user.username, user.password)
		}
	}
}

func TestDelete(t *testing.T) {
	for _, user := range userList {
		u, err := NewUserFromUsername(user.username)
		if err != nil {
			t.Fatalf("error in delete account %s", err)
		}
		deleted, err := u.Delete()
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		if !deleted {
			t.Fatalf("not delete %s", user.username)
		}
	}
}
