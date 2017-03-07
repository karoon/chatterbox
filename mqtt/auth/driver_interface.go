package auth

// Driver interface
type Driver interface {
	Register(u *User)
	CheckUsername(username string) (exist bool, err error)
	GetByUsername(username string) (u *User, err error)
	MakeID() string
	DeleteByUsername(username string) (deleted bool, err error)
}

type ModelDriver struct {
	Driver
}
