package mqtt

import (
	"net"

	log "github.com/cihub/seelog"
)

// Handle CONNECT
func Handle_(mqtt *Mqtt, conn *net.Conn, client **ClientRep) {
	log.Debugf("Hanling CONNECT, client id:(%s)", "")
}
