package auth

import (
	"chatterbox/boxconfig"
	"chatterbox/boxq/types"

	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

// var authModelDriver = "redis"

var authModelDriver = AuthDriverMongodb

// ErrUsernameExist username exist error
var ErrUsernameExist = errors.New("username already exist")

func SetDriver(authDriver types.AuthDriverType) {
	authModelDriver = authDriver
}

// User struct
type User struct {
	ID        string   `json:"id" bson:"_id"`
	Username  string   `json:"username" bson:"username"`
	Password  string   `json:"password" bson:"password"`
	Publish   []string `bson:"publish"`
	Subscribe []string `bson:"subscribe"`
	PubSub    []string `bson:"pubsub"`

	driver AuthDriver
}

func NewUserHandler() *User {
	u := &User{}

	switch boxconfig.NewConfigHandler().Auth.Driver {
	case AuthDriverRedis:
		u.driver = AuthModelDriver{newRedisDriver()}
	case AuthDriverMongodb:
		u.driver = AuthModelDriver{newMongoDriver()}
	}

	return u
}

func (u *User) UsernameExists() (exist bool, err error) {
	return u.driver.CheckUsername(u.Username)
}

func (u *User) Register() (user *User, err error) {
	ex, err := u.UsernameExists()
	if err != nil {
		return
	}
	if ex {
		return nil, ErrUsernameExist
	}

	u.driver.Register(u)

	return u, nil
}

func NewUserFromUsername(username string) (u *User, err error) {
	return NewUserHandler().driver.GetByUsername(username)
}

func (u *User) MakePasswordHash(p string) string {

	h := sha256.New()
	io.WriteString(h, p)

	pwmd5 := fmt.Sprintf("%x", h.Sum(nil))

	salt1 := "$2a$10$b4MO31Gyjs.qjI5f7JTzIOTrJ071ixZxl0mW.3WMJv1Kw2PGc7ZNe"
	salt2 := "$2a$10$ZiObC.2HCeV0VcAEdgADGeTL/7fBZUo/HokSwA4z8uUb8UF4jPWRO"

	// salt1 + username + salt2 + MD5 splicing
	io.WriteString(h, salt1)
	// io.WriteString(h, "abc")
	io.WriteString(h, salt2)
	io.WriteString(h, pwmd5)

	last := fmt.Sprintf("%x", h.Sum(nil))
	return last
}

func (u *User) PassHash() string {
	h := u.MakePasswordHash(u.Password)
	return string(h)
}

func (u *User) SetPassword(pasword string) *User {
	u.Password = pasword
	return u
}

func (u *User) SetUsername(username string) *User {
	u.Username = username
	return u
}

func (u *User) Login() (bool, error) {
	checkU, err := NewUserFromUsername(u.Username)
	if err != nil {
		return authDeny.Bool(), err
	}

	if u.PassHash() == checkU.Password {
		return authAllow.Bool(), nil
	}

	return boxconfig.NewConfigHandler().Auth.AuthNoMatch.Bool(), nil
}

func (u *User) DeleteByUsername() (deleted bool, err error) {
	return u.driver.DeleteByUsername(u.Username)
}

func (u *User) CheckACL(clientID, topic, acltype string) bool {
	actType, noMatch := u.driver.CheckACL(clientID, topic, acltype)
	if noMatch {
		actType = boxconfig.NewConfigHandler().Auth.ACLNoMatch.Bool()
	}
	return actType
}

func (u *User) SetACL(clientID, topic, acltype string) {
	u.driver.SetACL(clientID, topic, acltype)
}

func (u *User) RemoveACL(clientID, topic string) {
	// u.driver.RemoveACL(clientID, topic)
}
