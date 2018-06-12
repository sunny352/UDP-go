package network

import (
	"errors"
	"net"
	"time"
)

type connectWithListener struct {
	msgChan  chan []byte
	addr     *net.UDPAddr
	listener *Listener

	lastReceiveTime time.Time

	count int
}

func (self *connectWithListener) Read(b []byte) (n int, err error) {
	msg, err := self.Receive()
	if nil != err {
		return 0, err
	}
	copy(b, msg)
	return len(msg), nil
}

func (self *connectWithListener) Receive() ([]byte, error) {
	msg, ok := <-self.msgChan
	if !ok {
		return nil, errors.New("channel is closed")
	}
	self.count++
	return msg, nil
}

func (self *connectWithListener) Write(b []byte) (n int, err error) {
	return self.listener.send(b, self.addr)
}

func (self *connectWithListener) Close() error {
	close(self.msgChan)
	return self.listener.delete(self)
}

func (self *connectWithListener) LocalAddr() net.Addr {
	return self.listener.Addr()
}

func (self *connectWithListener) RemoteAddr() net.Addr {
	return self.RemoteUDPAddr()
}

func (self *connectWithListener) RemoteUDPAddr() *net.UDPAddr {
	return self.addr
}

func (self *connectWithListener) SetDeadline(t time.Time) error {
	return nil
}

func (self *connectWithListener) SetReadDeadline(t time.Time) error {
	return nil
}

func (self *connectWithListener) SetWriteDeadline(t time.Time) error {
	return nil
}
func (self *connectWithListener) push(msgBytes []byte) error {
	self.msgChan <- msgBytes
	self.lastReceiveTime = time.Now()
	return nil
}
func (self *connectWithListener) isTimeout() bool {
	return time.Now().Sub(self.lastReceiveTime) > time.Second*30
}

func CreateConnectWithListener(addr *net.UDPAddr, listener *Listener) (*connectWithListener, error) {
	connect := connectWithListener{
		msgChan:         make(chan []byte, 1024),
		addr:            addr,
		listener:        listener,
		lastReceiveTime: time.Now(),
	}
	return &connect, nil
}
