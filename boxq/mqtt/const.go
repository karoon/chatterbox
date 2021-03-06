package mqtt

// Types
const (
	CONNECT = uint8(iota + 1)
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

const (
	ACCEPTED = uint8(iota)
	UNACCEPTABLE_PROTOCOL_VERSION
	IDENTIFIER_REJECTED
	SERVER_UNAVAILABLE
	BAD_USERNAME_OR_PASSWORD
	NOT_AUTHORIZED
)

const (
	ClientIDLimit = 128
)

const (
	SEND_WILL = uint8(iota)
	DONT_SEND_WILL
)

const (
	PENDING_PUB = uint8(iota + 1)
	PENDING_ACK
)
