package network

import (
	"net"
	"time"
)

type connectWithoutListener struct {
	addr    *net.UDPAddr
	connect *net.UDPConn
}

func (self *connectWithoutListener) Read(b []byte) (n int, err error) {
	return self.connect.Read(b)
}

func (self *connectWithoutListener) Receive() ([]byte, error) {
	msgBytes := make([]byte, 2048)
	n, err := self.connect.Read(msgBytes)
	if nil != err {
		return nil, err
	}
	return msgBytes[:n], nil
}

func (self *connectWithoutListener) Write(b []byte) (n int, err error) {
	return self.connect.Write(b)
}

func (self *connectWithoutListener) Close() error {
	return self.connect.Close()
}

func (self *connectWithoutListener) LocalAddr() net.Addr {
	return self.connect.LocalAddr()
}

func (self *connectWithoutListener) RemoteAddr() net.Addr {
	return self.RemoteAddr()
}

func (self *connectWithoutListener) RemoteUDPAddr() *net.UDPAddr {
	return self.addr
}

func (self *connectWithoutListener) SetDeadline(t time.Time) error {
	return nil
}

func (self *connectWithoutListener) SetReadDeadline(t time.Time) error {
	return nil
}

func (self *connectWithoutListener) SetWriteDeadline(t time.Time) error {
	return nil
}

func CreateConnectWithoutListener(addr *net.UDPAddr, udp *net.UDPConn) (*connectWithoutListener, error) {
	connect := connectWithoutListener{
		addr:    addr,
		connect: udp,
	}
	return &connect, nil
}
