package messaging

import "log"

const (
	packetMessage = iota + 1

	// Actions packets
	packetTyping
	packetSeenMessage
	packetVideoUploading
	packetPhotoUploading
	packetFileUploading

	// Group actions packets
	packetGroupCreate
	packetGroupAddUser
	packetGroupDeleteUser // kick
	packetGroupLeftUser
	packetGroupChangePhoto
	packetGroupDeletePhoto
	packetGroupChangeTitle
	packetGroupDeleted

	packetChatHistory

	//
	packetCheckName
	packetBlockMemeber
	packetUnblockMember
	packetGetChatInfo
)

type BasePacketMessage struct {
	Type                  int      `json:"type"`
	Message               *Message `json:"message,omitempty"`
	GroupChatCreatedUsers []*User  `json:"members,omitempty"`
	BlockMember           *User    `json:"block_member,omitempty"`
}

func (p *BasePacketMessage) Process() {
	switch p.Type {
	case packetMessage:
		log.Println("we have new message")
	}
}

func NewPacketTextMessage() BasePacketMessage {
	pm := BasePacketMessage{
		Type: packetMessage,
		// Message:{

		// }
	}
}
