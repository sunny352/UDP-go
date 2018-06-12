package network

import (
	"net"
)

type Connect interface {
	net.Conn
	RemoteUDPAddr() *net.UDPAddr
	Receive() ([]byte, error)
}
