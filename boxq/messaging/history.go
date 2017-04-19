package messaging

const (
	chatTypePrivate = iota + 1
	chatTypeSecret
	chatTypeGroup
)

type History struct {
	ID           int // auto increment
	UserID       int // vahid by id 1
	Type         int // type of chat
	TargetUserID int // chat id
}
