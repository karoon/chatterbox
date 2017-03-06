package auth

import (
	"chatterbox/mqtt/connections"
	"errors"

	log "github.com/cihub/seelog"
	"github.com/pborman/uuid"
)

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

// ErrUsernameExist username exist error
var ErrUsernameExist = errors.New("username already exist")

type User struct {
	userKey  string
	id       string
	username string
	password Password
}

func NewUserHandler() *User {
	return &User{}
}

func NewUserFromUsername(username string) (u *User, err error) {
	u = &User{}
	u.SetUsername(username)
	rclient := connections.GetRedisClient()
	smap, err := rclient.HGetAll(u.userKey).Result()
	if err != nil {
		return nil, err
	}

	u.id = smap[authFieldID]
	u.password = Password(smap[authFieldPassword])

	return u, nil
}

func (u *User) SetPassword(pasword string) *User {
	u.password = Password(pasword)
	return u
}

func (u *User) setKey() {
	u.userKey = authPrefix + u.username
}

func (u *User) SetUsername(username string) *User {
	u.username = username
	u.setKey()
	return u
}

func (u *User) makeID() {
	u.id = uuid.New()
}

func (u *User) Login() (bool, error) {
	checkU, err := NewUserFromUsername(u.username)
	if err != nil {
		return authDeny, err
	}

	if u.password.Hash() == checkU.password.String() {
		return authAllow, nil
	}

	return authDeny, nil
}

func (u *User) Register() (user *User, err error) {
	ex, err := u.UsernameExists()
	if err != nil {
		return
	}
	if ex {
		return nil, ErrUsernameExist
	}
	rclient := connections.GetRedisClient()

	u.makeID()

	fields := make(map[string]string, 0)
	fields[authFieldPassword] = u.password.Hash()
	fields[authFieldUsername] = u.username
	fields[authFieldID] = u.id

	rclient.HMSet(u.userKey, fields).Result()
	return u, nil
}

func (u *User) Delete() (deleted bool, err error) {
	rclient := connections.GetRedisClient()
	_, err = rclient.Del(u.userKey).Result()
	if err != nil {
		log.Debugf("%s", err.Error())
		return false, err
	}
	return true, nil
}

func (u *User) UsernameExists() (exist bool, err error) {
	exist, err = connections.GetRedisClient().Exists(u.userKey).Result()
	if err != nil {
		log.Debugf("%s", err.Error())
		return false, err
	}
	return
}
