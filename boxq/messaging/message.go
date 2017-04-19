package messaging

const (
	messageTypeText = iota + 100
	messageTypePhoto
	messageTypeVideo
	messageTypeLocation
	messageTypeVoice
	messageTypeContact
	messageTypeFile
)

type Message struct {
	ID               int64     `json:"id"`
	Type             int       `json:"type"`
	Chat             Chat      `json:"chat"`
	From             User      `json:"from_user"`
	Date             int       `json:"date"`
	Text             string    `json:"text"`
	Video            *Video    `json:"video"`
	Photo            *Photo    `json:"photo"`
	Contact          *Contact  `json:"contact"`
	Location         *Location `json:"location"`
	Voice            *Voice    `json:"voice"`
	NewChatMember    *User     `json:"new_chat_member"`
	LeftChatMember   *User     `json:"left_chat_member"`
	RemoveChatMember *User     `json:"remove_chat_member"`
	NewChatTitle     string    `json:"new_chat_title"`
	NewChatPhoto     *[]Photo  `json:"new_chat_photo"`
	DeleteChatPhoto  bool      `json:"delete_chat_photo"`
	GroupChatCreated bool      `json:"group_chat_created"`
}

func NewTextMessage() Message {
	m := Message{
		ID:    1,
		Type:  messageTypeText,
		Topic: "single:vahid@amir",
	}
}
