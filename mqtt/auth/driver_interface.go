package auth

// AuthDriver interface
type AuthDriver interface {
	Register(u *User)
	CheckUsername(username string) (exist bool, err error)
	GetByUsername(username string) (u *User, err error)
	MakeID() string
	DeleteByUsername(username string) (deleted bool, err error)
	/* ACL */
	CheckACL(clientID, topic, acltype string) bool
	SetACL(clientID, topic, acltype string)
	RemoveACL(clientID, topic string)
}

type AuthModelDriver struct {
	AuthDriver
}
