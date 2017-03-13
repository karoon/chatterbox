package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	// huin "github.com/huin/mqtt"
	"golang.org/x/net/websocket"

	"github.com/cihub/seelog"

	"chatterbox/boxconfig"
	"chatterbox/boxq/mqtt"
	"chatterbox/boxq/packet"
	// "chatterbox/boxq/proto"
)

var gDebug = flag.Bool("d", false, "enable debugging log")
var gPort = flag.Int("p", 1883, "port of the broker to listen")
var gRedisPort = flag.Int("r", 6379, "port of the broker to listen")

var chatConfig *boxconfig.Configuration

func setupLogging() {
	level := "info"
	if *gDebug == true {
		level = "debug"
	}
	config := fmt.Sprintf(`
<seelog type="sync" minlevel="%s">
	<outputs formatid="main">
		<console/>
	</outputs>
	<formats>
		<format id="main" format="%%Date %%Time [%%LEVEL] %%File|%%FuncShort|%%Line: %%Msg%%n"/>
	</formats>
</seelog>`, level)

	logger, err := seelog.LoggerFromConfigAsBytes([]byte(config))

	if err != nil {
		fmt.Println("Failed to config logging:", err)
		os.Exit(1)
	}

	seelog.ReplaceLogger(logger)

	seelog.Info("Logging config is successful")
}

func tcp1883() {
	finish := make(chan bool)
	seelog.Debugf("Chatterbox kicking off, listening localhost:%d", *gPort)

	link, _ := net.Listen("tcp", fmt.Sprintf(":%d", *gPort))
	defer link.Close()

	go func() {
		for {
			conn, err := link.Accept()
			if err != nil {
				continue
			}

			c := mqtt.NewConnFromNetConn(&conn)
			seelog.Debugf("format %s", c)

			go mqtt.HandleConnection(&c)
		}
	}()
	<-finish
}

func main() {
	finish := make(chan bool)
	flag.Parse()
	setupLogging()
	chatConfig = boxconfig.NewConfigHandler()
	mqtt.RecoverFromRedis()

	// go tcp8883()
	go tcp1883()
	// go ws()

	http.Handle("/", websocket.Handler(wshandler))

	err := http.ListenAndServe(":8083", nil)

	if err != nil {
		fmt.Println("ListenAndserve:", err)
		return
	}

	<-finish
}

func wshandler(ws *websocket.Conn) {
	ws.PayloadType = websocket.BinaryFrame

	// bws := bufio.NewReadWriter(bufio.NewReader(ws), bufio.NewWriter(ws))

	for {
		// var hdr proto.Header
		// msgType, _, _ := hdr.Decode(bws)

		// b, err := bws.ReadBytes(delim)
		var data []byte
		websocket.Message.Receive(ws, &data)
		l, mt := packet.DetectPacket(data)

		seelog.Debug(l, mt)

		pkt2, err := mt.New()
		if err != nil {
			seelog.Debug(err.Error()) // packet type is invalid
			return
		}

		// Decode packet.
		_, err = pkt2.Decode(data)
		if err != nil {
			seelog.Debug(err.Error()) // there was an error while decoding
			return
		}

		switch pkt2.Type() {
		case packet.CONNECT:
			c := pkt2.(*packet.ConnectPacket)
			seelog.Debug(c.Username, c.Password)
			packet.NewConnackPacket().Encode(data)
			websocket.Message.Send(ws, data)
		case packet.DISCONNECT:
			packet.NewDisconnectPacket().Encode(data)
			websocket.Message.Send(ws, data)
		case packet.PUBLISH:
			packet.NewPubackPacket().Encode(data)
			websocket.Message.Send(ws, data)
		}

		// msg, err := proto.DecodeOneMessage(bws, nil)

		// seelog.Debug("webs->", msg, bws)
		// if err != nil {
		// 	seelog.Debug(err.Error())
		// 	seelog.Debug("close connection")
		// 	ws.Close()
		// 	return
		// }

		// switch msgType {
		// case proto.MsgConnect:
		// 	ca, err := proto.NewMessage(proto.MsgConnAck)
		// 	if err != nil {
		// 		seelog.Debug(err.Error())
		// 	}
		// 	wbuffer := new(bytes.Buffer)
		// 	ca.Encode(wbuffer)
		// 	bws.Write(wbuffer.Bytes())

		// 	bws.Flush()
		// }

	}
}

// ssl implementation
// func tcp8883() {
// 	finish := make(chan bool)

// 	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
// 	if err != nil {
// 		seelog.Debugf("%s", err)
// 		return
// 	}

// 	config := &tls.Config{Certificates: []tls.Certificate{cer}}
// 	ln, err := tls.Listen("tcp", ":8883", config)
// 	if err != nil {
// 		seelog.Debugf("%s", err)
// 		return
// 	}
// 	defer ln.Close()

// 	go func() {
// 		for {
// 			conn, err := ln.Accept()
// 			if err != nil {
// 				seelog.Debugf("%s", err)
// 				continue
// 			}
// 			go mqtt.HandleConnection(&conn)
// 		}
// 	}()
// 	<-finish
// }
