package algorithm

import (
	"net"
	"time"

	"container/list"
	"sync"
)

type ConnectMock struct {
	dropPercent float32

	minDelay int
	maxDelay int

	readTimer  *time.Timer
	writeTimer *time.Timer

	msgLocker sync.Mutex
	msgList   *list.List

	msgChan chan []byte
}

func (self *ConnectMock) Read(b []byte) (n int, err error) {
	msg := <-self.msgChan
	copy(b, msg)
	return len(msg), nil

	//for {
	//	func() {
	//		self.msgLocker.Lock()
	//		defer self.msgLocker.Unlock()
	//		currentTime := time.Now()
	//		for e := self.msgList.Front(); nil != e; e = e.Next() {
	//			info := e.Value.(*delayPackage)
	//			if info.postTime.Before(currentTime) {
	//				copy(b, info.msg)
	//				self.msgList.Remove(e)
	//				n = len(info.msg)
	//				//log.Println(self.msgList.Len())
	//				break
	//			}
	//		}
	//	}()
	//	if n > 0 {
	//		return n, nil
	//	}
	//	//time.Sleep(time.Millisecond)
	//	runtime.Gosched()
	//}
}

func (self *ConnectMock) Write(b []byte) (n int, err error) {
	self.msgChan <- b
	return len(b), nil
	//if self.dropPercent > 0 {
	//	if rand.Float32() < self.dropPercent {
	//		return len(b), nil
	//	}
	//}
	//var delay time.Duration
	//if self.minDelay == self.maxDelay {
	//	delay = time.Duration(self.minDelay)
	//} else {
	//	delay = time.Duration(rand.Intn(self.maxDelay-self.minDelay) + self.minDelay)
	//}
	//self.msgLocker.Lock()
	//defer self.msgLocker.Unlock()
	//self.msgList.PushBack(&delayPackage{
	//	postTime: time.Now().Add(delay),
	//	msg:      b,
	//})
	//return len(b), nil
}

func (self *ConnectMock) Close() error {
	return nil
}

func (self *ConnectMock) LocalAddr() net.Addr {
	return nil
}

func (self *ConnectMock) RemoteAddr() net.Addr {
	return nil
}

func (self *ConnectMock) SetDeadline(t time.Time) error {
	self.SetReadDeadline(t)
	self.SetWriteDeadline(t)
	return nil
}

func (self *ConnectMock) SetReadDeadline(t time.Time) error {
	if nil == self.readTimer {
		self.readTimer = time.NewTimer(t.Sub(time.Now()))
	} else {
		self.readTimer.Reset(t.Sub(time.Now()))
	}
	return nil
}

func (self *ConnectMock) SetWriteDeadline(t time.Time) error {
	if nil == self.writeTimer {
		self.writeTimer = time.NewTimer(t.Sub(time.Now()))
	} else {
		self.writeTimer.Reset(t.Sub(time.Now()))
	}
	return nil
}

type delayPackage struct {
	postTime time.Time
	msg      []byte
}

func CreateConnectMock(dropPercent float32, minDelay, maxDelay int) *ConnectMock {
	connect := &ConnectMock{
		dropPercent: dropPercent,
		minDelay:    minDelay,
		maxDelay:    maxDelay,
		msgList:     list.New(),
		msgChan:     make(chan []byte, 1024),
	}
	return connect
}
