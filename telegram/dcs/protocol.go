package dcs

import (
	"net"

	"github.com/owpengram/owpengram-td/transport"
)

type protocol interface {
	Handshake(conn net.Conn) (transport.Conn, error)
}
