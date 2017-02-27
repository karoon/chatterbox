package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime/debug"

	log "github.com/cihub/seelog"

	"chatterbox/mqtt"
)

type CmdFunc func(mqtt *mqtt.Mqtt, conn *net.Conn, client **mqtt.ClientRep)

var gDebug = flag.Bool("d", false, "enable debugging log")
var gPort = flag.Int("p", 1883, "port of the broker to listen")
var gRedisPort = flag.Int("r", 6379, "port of the broker to listen")

var gCmdRoute = map[uint8]CmdFunc{
	mqtt.CONNECT:     mqtt.HandleConnect,
	mqtt.PUBLISH:     mqtt.HandlePublish,
	mqtt.SUBSCRIBE:   mqtt.HandleSubscribe,
	mqtt.UNSUBSCRIBE: mqtt.HandleUnsubscribe,
	mqtt.PINGREQ:     mqtt.HandlePingreq,
	mqtt.DISCONNECT:  mqtt.HandleDisconnect,
	mqtt.PUBACK:      mqtt.HandlePuback,
}

func handleConnection(conn *net.Conn) {
	remoteAddr := (*conn).RemoteAddr()
	var client *mqtt.ClientRep

	defer func() {
		log.Debug("executing defered func in handleConnection")
		if r := recover(); r != nil {
			log.Debugf("got panic:(%s) will close connection from %s:%s", r, remoteAddr.Network(), remoteAddr.String())
			debug.PrintStack()
		}
		if client != nil {
			mqtt.ForceDisconnect(client, mqtt.GlobalClientsLock, mqtt.SEND_WILL)
		}
		(*conn).Close()
	}()

	connStr := fmt.Sprintf("%s:%s", string(remoteAddr.Network()), remoteAddr.String())
	log.Debug("Got new conection", connStr)
	for {
		// Read fixed header
		fixedHeader, body := mqtt.ReadCompleteCommand(conn)
		if fixedHeader == nil {
			log.Debug(connStr, "reading header returned nil, will disconnect")
			return
		}

		mqttParsed, err := mqtt.DecodeAfterFixedHeader(fixedHeader, body)
		if err != nil {
			log.Debug(connStr, "read command body failed:", err.Error())
		}

		var clientID string
		if client == nil {
			clientID = ""
		} else {
			clientID = client.ClientID
		}
		log.Debugf("Got request: %s from %s", mqtt.MessageTypeStr(fixedHeader.MessageType), clientID)
		proc, found := gCmdRoute[fixedHeader.MessageType]
		if !found {
			log.Debugf("Handler func not found for message type: %d(%s)",
				fixedHeader.MessageType, mqtt.MessageTypeStr(fixedHeader.MessageType))
			return
		}
		proc(mqttParsed, conn, &client)
	}
}

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

	logger, err := log.LoggerFromConfigAsBytes([]byte(config))

	if err != nil {
		fmt.Println("Failed to config logging:", err)
		os.Exit(1)
	}

	log.ReplaceLogger(logger)

	log.Info("Logging config is successful")
}

func main() {
	flag.Parse()

	setupLogging()

	mqtt.RecoverFromRedis()

	log.Debugf("Gossipd kicking off, listening localhost:%d", *gPort)

	link, _ := net.Listen("tcp", fmt.Sprintf(":%d", *gPort))
	defer link.Close()

	for {
		conn, err := link.Accept()
		if err != nil {
			continue
		}
		go handleConnection(&conn)
	}
}
