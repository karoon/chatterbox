package mqtt

import (
	"fmt"
	"io"
	"net"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/cihub/seelog"
)

// Network connection types
const (
	NetConnection int = iota + 1
	WSConnections
)

// Conn struct
type Conn struct {
	netConn  *net.Conn
	wsConn   *websocket.Conn
	typeConn int
	*sync.Mutex
}

func NewConnFromNetConn(netConn *net.Conn) Conn {
	c := Conn{}
	c.netConn = netConn
	c.typeConn = NetConnection
	return c
}

func NewConnFromWebSocket(wsConn *websocket.Conn) Conn {
	c := Conn{}
	c.wsConn = wsConn
	c.typeConn = WSConnections
	return c
}

func (c *Conn) getConnection() interface{} {
	return c.netConn
}

func (c *Conn) Write(bytes []byte) {
	(*c.netConn).Write(bytes)
}

// RemoteAddr function
func (c *Conn) RemoteAddr() net.Addr {
	switch c.typeConn {
	case NetConnection:
		return (*c.netConn).RemoteAddr()
	case WSConnections:
		return (c.wsConn).RemoteAddr()
	}
	return nil
}

// Close connection
func (c *Conn) Close() {
	(*c.netConn).Close()
}

// ReadFixedHeader docs
// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718020
func (c *Conn) readFixedHeader() *FixedHeader {
	var buf = make([]byte, 2)
	seelog.Debug("here ")
	switch c.typeConn {
	case NetConnection:
		n, _ := io.ReadFull(*c.netConn, buf)
		if n != len(buf) {
			seelog.Debug("read header failed")
			return nil
		}
	case WSConnections:
		var msg []byte
		websocket.Message.Receive(c.wsConn, &msg)
		seelog.Debugf("wsconnection ", msg)
		// n, err := c.wsConn.Read(msg)
		// if err != nil {
		// 	seelog.Debugf("error %s", err)
		// }
		// seelog.Debugf("%d", len(buf))
		// if n != len(buf) {
		// 	seelog.Debug("read header ws failed")
		// 	return nil
		// }
	}

	seelog.Debug(buf)

	byte1 := buf[0]
	header := new(FixedHeader)
	header.MessageType = uint8(byte1 & 0xF0 >> 4)
	header.DupFlag = byte1&0x08 > 0
	header.QosLevel = uint8(byte1 & 0x06 >> 1)
	header.Retain = byte1&0x01 > 0

	byte2 := buf[1]
	header.Length = c.decodeVarLength(byte2, c.netConn)
	return header
}

func (c *Conn) readCompleteCommand() (*FixedHeader, []byte) {
	fixedHeader := c.readFixedHeader()
	if fixedHeader == nil {
		seelog.Debug("failed to read fixed header")
		return nil, make([]byte, 0)
	}
	seelog.Debug("start read command")
	length := fixedHeader.Length
	buf := make([]byte, length)
	n, _ := io.ReadFull(*c.netConn, buf)
	if uint32(n) != length {
		panic(fmt.Sprintf("failed to read %d bytes specified in fixed header, only %d read", length, n))
	}
	seelog.Debugf("Complete command(%s) read into buffer", messageTypeStr(fixedHeader.MessageType))

	return fixedHeader, buf
}

func (c *Conn) decodeVarLength(cur byte, conn *net.Conn) uint32 {
	length := uint32(0)
	multi := uint32(1)

	for {
		length += multi * uint32(cur&0x7f)
		if cur&0x80 == 0 {
			break
		}
		buf := make([]byte, 1)
		n, _ := io.ReadFull(*conn, buf)
		if n != 1 {
			panic("failed to read variable length in MQTT header")
		}
		cur = buf[0]
		multi *= 128
	}

	return length
}
