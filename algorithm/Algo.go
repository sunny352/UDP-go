package algorithm

import (
	"io"
	"net"

	"container/list"

	"sync/atomic"

	"errors"

	"time"

	"sync"

	"sort"

	"github.com/sunny352/UDP-go/Utils/ByteOpt"
	"github.com/sunny352/UDP-go/Utils/SLog"
	"github.com/sunny352/UDP-go/Utils/Safe"
)

const (
	NextSendDuration = time.Millisecond * 500
)

type Algo interface {
	Init(connect net.Conn, owner uint32, packLength uint16)
	io.Closer
	Send(msg []byte) error
	Receive() ([]byte, error)

	Remain() int
}

func CreateAlgo(connect net.Conn, owner uint32, packLength uint16) Algo {
	algo := &algoImpl{}
	algo.Init(connect, owner, packLength)
	return algo
}

type algoImpl struct {
	connect    net.Conn
	owner      uint32
	packLength uint16

	sendList   *list.List
	sendLocker sync.RWMutex
	index      uint32
	key        uint32

	receiveChan chan []byte
	receiveMap  map[uint32]*Combiner

	confirmChan chan uint32

	isRunning int32
}

func (self *algoImpl) Init(connect net.Conn, owner uint32, packLength uint16) {
	self.connect = connect
	self.owner = owner
	self.packLength = packLength

	self.sendList = list.New()
	self.receiveChan = make(chan []byte, 1024)
	self.receiveMap = make(map[uint32]*Combiner)

	self.confirmChan = make(chan uint32, 1024)

	Safe.Go(self.run)
}

func (self *algoImpl) Close() error {
	if 0 != atomic.LoadInt32(&self.isRunning) {
		return nil
	}
	atomic.AddInt32(&self.isRunning, 1)
	closeBytes, _ := PackCloseBytes(&ClosePackage{
		Owner: self.owner,
	})
	self.connect.Write(closeBytes)
	self.connect.Close()
	self.connect = nil
	self.sendList = nil
	self.receiveMap = nil
	close(self.receiveChan)
	close(self.confirmChan)
	return nil
}

type sendInfo struct {
	Index    uint32
	Msg      []byte
	NextSend time.Time
}

func (self *algoImpl) Send(msg []byte) error {
	if 0 != atomic.LoadInt32(&self.isRunning) {
		return errors.New("closed")
	}
	packList, index, err := SeparatePackage(msg, self.owner, self.index, self.key, self.packLength)
	self.index = index
	self.key++
	if nil != err {
		return err
	}
	currentTime := time.Now()
	nextSendTime := currentTime.Add(NextSendDuration)
	func() {
		bytesList := make([]*sendInfo, len(packList))
		for key, value := range packList {
			pack, err := PackDataBytes(&value)
			if nil != err {
				bytesList = nil
				break
			}
			bytesList[key] = &sendInfo{
				Index:    value.Index,
				Msg:      pack,
				NextSend: nextSendTime,
			}
		}
		if nil != bytesList {
			func() {
				self.sendLocker.Lock()
				defer self.sendLocker.Unlock()
				for _, value := range bytesList {
					self.sendList.PushBack(value)
				}
			}()
			for _, value := range bytesList {
				_, err = self.connect.Write(value.Msg)
				if nil != err {
					SLog.E(err)
					break
				}
			}
		}
	}()
	return nil
}
func (self *algoImpl) Receive() ([]byte, error) {
	if 0 != atomic.LoadInt32(&self.isRunning) {
		return nil, errors.New("closed")
	}
	msg, ok := <-self.receiveChan
	if !ok {
		return nil, errors.New("channel is closed")
	}
	return msg, nil
}

func (self *algoImpl) Remain() int {
	self.sendLocker.RLock()
	defer self.sendLocker.RUnlock()
	return self.sendList.Len()
}

func (self *algoImpl) processDataPackage(buffer []byte) error {
	var dataPackage DataPackage
	err := UnPackData(&dataPackage, buffer)
	if nil != err {
		SLog.E(err)
		return err
	}
	self.confirmChan <- dataPackage.Index
	combiner, ok := self.receiveMap[dataPackage.Key]
	if !ok {
		combiner = &Combiner{}
		combiner.Init(dataPackage.KCount)
		self.receiveMap[dataPackage.Key] = combiner
	}
	if combiner.Push(&dataPackage) {
		if nil != self.receiveChan {
			self.receiveChan <- combiner.Bytes()
		}
		delete(self.receiveMap, dataPackage.Key)
	}
	return nil
}

func (self *algoImpl) processConfirmPackage(buffer []byte) error {
	var confirmPackage ConfirmPackage
	err := UnPackConfirm(&confirmPackage, buffer)
	if nil != err {
		SLog.E(err)
		return err
	}
	confirmPackage.IndexArray.Sort()

	self.sendLocker.Lock()
	defer self.sendLocker.Unlock()

	for elem := self.sendList.Front(); nil != elem; {
		current := elem
		elem = elem.Next()

		info := current.Value.(*sendInfo)

		index := sort.Search(len(confirmPackage.IndexArray), func(i int) bool {
			return confirmPackage.IndexArray[i] >= info.Index
		})
		if index >= len(confirmPackage.IndexArray) {
			continue
		}
		if confirmPackage.IndexArray[index] == info.Index {
			self.sendList.Remove(current)
		}
	}
	return nil
}

func (self *algoImpl) sendConfirm(indexArray []uint32) error {
	packBytes, err := PackConfirmBytes(&ConfirmPackage{
		Owner:      self.owner,
		IndexArray: indexArray,
	})
	if nil != err {
		SLog.E(err)
		return err
	}
	self.connect.Write(packBytes)
	return nil
}

func (self *algoImpl) resendMsg() {
	currentTime := time.Now()
	self.sendLocker.RLock()
	defer self.sendLocker.RUnlock()

	for e := self.sendList.Front(); nil != e; e = e.Next() {
		info := e.Value.(*sendInfo)
		if info.NextSend.Before(currentTime) {
			self.connect.Write(info.Msg)
			info.NextSend = info.NextSend.Add(NextSendDuration)
		}
	}
}

type confirmArrayHelper struct {
	data   []uint32
	length int
}

func (self *confirmArrayHelper) Reset() {
	self.length = 0
}
func (self *confirmArrayHelper) Add(value uint32) {
	self.data[self.length] = value
	self.length++
}
func (self *confirmArrayHelper) IsFull() bool {
	return self.length >= len(self.data)
}
func (self *confirmArrayHelper) IsEmpty() bool {
	return 0 == self.length
}
func (self *confirmArrayHelper) ToSlice() []uint32 {
	return self.data[:self.length]
}

func (self *algoImpl) run() {
	confirmDuration := time.Millisecond * 100 //每100ms强制检测发送一次确认消息
	confirmTimer := time.NewTimer(confirmDuration)

	resendDuration := time.Millisecond * 10 //每10ms检测一次消息重传
	resendTimer := time.NewTimer(resendDuration)

	tickDuration := time.Second * 10 //每10s发送一次连接包
	tickTimer := time.NewTimer(0)
	tickCount := int32(0)

	arrayHelper := confirmArrayHelper{data: make([]uint32, 128), length: 0}

	Safe.Go(func() {
		for 0 == atomic.LoadInt32(&self.isRunning) {
			buffer := make([]byte, 2048)
			n, err := self.connect.Read(buffer)
			if nil != err {
				SLog.E(err)
				break
			}
			var packType byte
			buffer = ByteOpt.Decode8u(buffer[:n], &packType)
			switch packType {
			case PackType_Data:
				self.processDataPackage(buffer)
			case PackType_Confim:
				self.processConfirmPackage(buffer)
			case PackType_Close:
				self.Close()
			case PackType_Tick:
				var tickPackage TickPackage
				UnPackTickPackage(&tickPackage, buffer)
				echoBytes, _ := PackEchoTickPackage(&EchoTickPackage{
					Owner: tickPackage.Owner,
					NSec:  tickPackage.NSec,
				})
				self.connect.Write(echoBytes)
			case PackType_EchoTick:
				atomic.StoreInt32(&tickCount, 0)
				var echoTickPackage EchoTickPackage
				UnPackEchoTickPackage(&echoTickPackage, buffer)
				pre := time.Unix(0, echoTickPackage.NSec)
				current := time.Now()
				SLog.D(time.Now().Sub(pre), pre, current)
			default:
			}
		}
	})
	for 0 == atomic.LoadInt32(&self.isRunning) {
		select {
		case index, ok := <-self.confirmChan:
			if ok {
				arrayHelper.Add(index)
				if arrayHelper.IsFull() {
					self.sendConfirm(arrayHelper.ToSlice())
					arrayHelper.Reset()
				}
			}
		case <-confirmTimer.C:
			if !arrayHelper.IsEmpty() {
				self.sendConfirm(arrayHelper.ToSlice())
				arrayHelper.Reset()
			}
			confirmTimer.Reset(confirmDuration)
		case <-resendTimer.C:
			self.resendMsg()
			resendTimer.Reset(resendDuration)
		case <-tickTimer.C:
			nowTick := atomic.AddInt32(&tickCount, 1)
			if nowTick > 3 {
				self.Close()
				SLog.D("Close")
			} else {
				current := time.Now()
				tickBytes, _ := PackTickPackage(&TickPackage{
					Owner: self.owner,
					NSec:  current.UnixNano(),
				})
				self.connect.Write(tickBytes)
				SLog.D(current, "Tick")
			}
			tickTimer.Reset(time.Duration(float32(tickDuration) * (1.0 / float32(nowTick))))
		}
	}
}
