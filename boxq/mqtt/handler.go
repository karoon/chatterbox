package mqtt

import (
	"fmt"
	"net"
	"runtime/debug"

	"github.com/cihub/seelog"
	// "github.com/luanjunyi/gossipd/mqtt"
)

type CmdFunc func(mqtt *Mqtt, conn *net.Conn, client **ClientRep)

var gCmdRoute = map[uint8]CmdFunc{
	CONNECT:     handleConnect,
	PUBLISH:     handlePublish,
	SUBSCRIBE:   handleSubscribe,
	UNSUBSCRIBE: handleUnsubscribe,
	PINGREQ:     handlePingreq,
	DISCONNECT:  handleDisconnect,
	PUBACK:      handlePuback,
	PUBREL:      handlePubrel,
}

// HandleConnection handle connection from network
func HandleConnection(conn *net.Conn) {
	remoteAddr := (*conn).RemoteAddr()
	var client *ClientRep

	defer func() {
		seelog.Debug("executing defered func in handleConnection")
		if r := recover(); r != nil {
			seelog.Debugf("got panic:(%s) will close connection from %s:%s", r, remoteAddr.Network(), remoteAddr.String())
			debug.PrintStack()
		}
		if client != nil {
			ForceDisconnect(client, GlobalClientsLock, SEND_WILL)
		}
		(*conn).Close()
	}()

	connStr := fmt.Sprintf("%s:%s", string(remoteAddr.Network()), remoteAddr.String())
	seelog.Debug("Got new conection ", connStr)
	for {
		// Read fixed header
		fixedHeader, body := readCompleteCommand(conn)
		if fixedHeader == nil {
			seelog.Debug(connStr, "reading header returned nil, will disconnect")
			return
		}

		seelog.Debugf("message type: %s ", messageTypeStr(fixedHeader.MessageType))

		mqttParsed, err := decodeAfterFixedHeader(fixedHeader, body)
		if err != nil {
			seelog.Debug(connStr, "read command body failed:", err.Error())
		}

		var clientID string
		if client == nil {
			clientID = ""
		} else {
			clientID = client.ClientID
		}

		seelog.Debugf("Got request: %s from %s", messageTypeStr(fixedHeader.MessageType), clientID)

		proc, found := gCmdRoute[fixedHeader.MessageType]
		if !found {
			seelog.Debugf("Handler func not found for message type: %d(%s)",
				fixedHeader.MessageType, messageTypeStr(fixedHeader.MessageType))
			return
		}
		proc(mqttParsed, conn, &client)
	}
}
