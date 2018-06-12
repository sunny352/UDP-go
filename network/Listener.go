package network

import (
	"errors"
	"net"
	"sync"

	"time"

	"github.com/sunny352/UDP-go/Utils/SLog"
	"github.com/sunny352/UDP-go/Utils/Safe"
)

type Listener struct {
	udp    *net.UDPConn
	locker sync.RWMutex
	udpMap map[string]*connectWithListener

	connectChan chan Connect

	clearTimer *time.Timer
}

func (self *Listener) Accept() (net.Conn, error) {
	return self.AcceptUDP()
}

func (self *Listener) AcceptUDP() (Connect, error) {
	connect, ok := <-self.connectChan
	if !ok {
		return nil, errors.New("channel is closed")
	}
	return connect, nil
}

func (self *Listener) Close() error {
	err := self.udp.Close()
	if nil != err {
		return err
	}

	close(self.connectChan)

	self.locker.Lock()
	defer self.locker.Unlock()
	for _, value := range self.udpMap {
		value.Close()
	}
	self.udpMap = nil
	return nil
}

func (self *Listener) Addr() net.Addr {
	return self.udp.LocalAddr()
}

func (self *Listener) run() {
	defer self.Close()
	count := 0
	for {
		buffer := make([]byte, 2048)
		n, udpAddr, err := self.udp.ReadFromUDP(buffer)
		if nil != err {
			SLog.E(err)
			break
		}
		count++
		connect := func() *connectWithListener {
			self.locker.Lock()
			defer self.locker.Unlock()
			connect, ok := self.udpMap[udpAddr.String()]
			if !ok {
				connect, err = CreateConnectWithListener(udpAddr, self)
				if nil != err {
					return nil
				}
				self.udpMap[udpAddr.String()] = connect
				self.connectChan <- connect
			}
			return connect
		}()
		if nil != connect {
			connect.push(buffer[:n])
		}
	}
}

func (self *Listener) loopClear() {
	for range self.clearTimer.C {
		for _, value := range self.udpMap {
			if value.isTimeout() {
				value.Close()
			}
		}
		self.clearTimer.Reset(time.Second)
	}
}

func (self *Listener) send(msgBytes []byte, addr *net.UDPAddr) (int, error) {
	return self.udp.WriteToUDP(msgBytes, addr)
}
func (self *Listener) delete(connect Connect) error {
	self.locker.Lock()
	defer self.locker.Unlock()
	delete(self.udpMap, connect.RemoteAddr().String())
	return nil
}

func Listen(network, address string) (*Listener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if nil != err {
		return nil, err
	}
	udp, err := net.ListenUDP(network, udpAddr)
	if nil != err {
		return nil, err
	}
	//udp.SetReadBuffer(1024 * 1024 * 1024)
	//udp.SetWriteBuffer(1024 * 1024 * 1024)
	listener := &Listener{
		udp:         udp,
		udpMap:      make(map[string]*connectWithListener),
		connectChan: make(chan Connect),
		clearTimer:  time.NewTimer(time.Second),
	}
	Safe.Go(listener.run)
	Safe.Go(listener.loopClear)
	return listener, nil
}
