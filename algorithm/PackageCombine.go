package algorithm

import (
	"github.com/sunny352/UDP-go/Utils/SLog"
)

type Combiner struct {
	Data   [][]byte
	Total  uint16
	Count  uint16
	Length int
}

func (self *Combiner) Init(total uint16) {
	if nil != self.Data {
		if len(self.Data) < int(total) {
			self.Data = make([][]byte, total)
		}
	} else {
		self.Data = make([][]byte, total)
	}
	self.Total = total
}
func (self *Combiner) Push(pack *DataPackage) bool {
	if self.Total <= pack.KIndex {
		SLog.W("false", self.Total, pack.KIndex)
		return false
	}
	if nil != self.Data[pack.KIndex] {
		return false
	}
	self.Data[pack.KIndex] = pack.Data
	self.Count++
	self.Length += int(pack.Length)
	return self.Count == self.Total
}

func (self *Combiner) Bytes() []byte {
	if self.Count != self.Total {
		return nil
	}
	buffer := make([]byte, self.Length)
	current := 0
	for _, value := range self.Data {
		length := len(value)
		copy(buffer[current:current+length], value)
		current += length
	}
	return buffer
}
func (self *Combiner) Reset() {
	self.Total = 0
	self.Count = 0
	for key := range self.Data {
		self.Data[key] = nil
	}
}
