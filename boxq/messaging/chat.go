package messaging

const (
	chatTypePrivate = iota + 1
	chatTypeSecret
	chatTypeGroup
)

type Chat struct {
	Code            string  `json:"code"`
	Type            int     `json:"type"`
	LastMessageDate int     `json:"last_message_date"`
	UnreadCount     int     `json:"unread_count"`
	LastMessage     Message `json:"last_message"`
}
