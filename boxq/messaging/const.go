package messaging

const (
	Message = iota + 100
)

const (
	StartChat = iota + 200
)

const (
	GroupInvite = iota + 300
	GroupKick
	GroupLeft
	GroupAddAdmin
)
