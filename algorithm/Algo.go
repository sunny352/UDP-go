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

	Safe.Go(self.runConfirm)
	Safe.Go(self.runReceive)
}

func (self *algoImpl) Close() error {
	if 0 != atomic.LoadInt32(&self.isRunning) {
		return nil
	}
	atomic.AddInt32(&self.isRunning, 1)
	close(self.receiveChan)
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

func (self *algoImpl) runReceive() {
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
			var dataPackage DataPackage
			err := UnPackData(&dataPackage, buffer)
			if nil != err {
				SLog.E(err)
				break
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
			break
		case PackType_Confim:
			var confirmPackage ConfirmPackage
			err := UnPackConfirm(&confirmPackage, buffer)
			if nil != err {
				SLog.E(err)
				break
			}
			confirmPackage.IndexArray.Sort()
			func() {
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
			}()
			break
		case PackType_Close:
			self.Close()
			break
		default:
			break
		}
	}
	SLog.D("End")
}

func (self *algoImpl) runConfirm() {
	duration := time.Millisecond * 100
	confirmTimer := time.NewTimer(time.Millisecond * 100)
	arrayLength := 128
	indexArray := make([]uint32, arrayLength)
	confirmMsg := ConfirmPackage{
		Owner:      self.owner,
		IndexArray: make([]uint32, arrayLength),
	}
	confirmIndex := 0
	sendConfirm := func(msg *ConfirmPackage) {
		packBytes, err := PackConfirmBytes(msg)
		if nil != err {
			SLog.E(err)
			return
		}
		self.connect.Write(packBytes)
	}
	for 0 == atomic.LoadInt32(&self.isRunning) {
		select {
		case index, ok := <-self.confirmChan:
			if ok {
				indexArray[confirmIndex] = index
				confirmIndex++
				if confirmIndex >= len(confirmMsg.IndexArray) {
					confirmMsg.IndexArray = indexArray[:confirmIndex]
					sendConfirm(&confirmMsg)
					for innerIndex := 0; innerIndex < len(indexArray); innerIndex++ {
						indexArray[innerIndex] = 0
					}
					confirmIndex = 0
				}
			}
		case <-confirmTimer.C:
			if confirmIndex > 0 {
				confirmMsg.IndexArray = indexArray[:confirmIndex]
				sendConfirm(&confirmMsg)
				for innerIndex := 0; innerIndex < len(indexArray); innerIndex++ {
					indexArray[innerIndex] = 0
				}
				confirmIndex = 0
			}
			confirmTimer.Reset(duration)
		default:
			currentTime := time.Now()
			func() {
				self.sendLocker.RLock()
				defer self.sendLocker.RUnlock()

				for e := self.sendList.Front(); nil != e; e = e.Next() {
					info := e.Value.(*sendInfo)
					if info.NextSend.Before(currentTime) {
						self.connect.Write(info.Msg)
						info.NextSend = info.NextSend.Add(NextSendDuration)
					}
				}
			}()
		}
	}
}
