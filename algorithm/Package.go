package algorithm

import (
	"sort"

	"github.com/sunny352/UDP-go/Utils/ByteOpt"
)

const (
	PackType_Data     byte = 0
	PackType_Confim        = 1
	PackType_Close         = 2
	PackType_Tick          = 4
	PackType_EchoTick      = 5
)

type DataPackage struct {
	Owner uint32
	//包序列
	Index uint32
	//消息key
	Key uint32
	//包数据长度
	Length uint16
	//消息的包序列
	KIndex uint16
	//消息的总包数
	KCount uint16
	//数据
	Data []byte
}

const (
	DataPackageHeadSize    = 4 + 4 + 4 + 2 + 2 + 2
	ConfirmPackageHeadSize = 4
)

func PackDataBytes(dataPackage *DataPackage) ([]byte, error) {
	buffer := make([]byte, 1+DataPackageHeadSize+len(dataPackage.Data))
	temp := buffer
	temp = ByteOpt.Encode8u(temp, PackType_Data)
	temp = ByteOpt.Encode32u(temp, dataPackage.Owner)
	temp = ByteOpt.Encode32u(temp, dataPackage.Index)
	temp = ByteOpt.Encode32u(temp, dataPackage.Key)
	temp = ByteOpt.Encode16u(temp, dataPackage.Length)
	temp = ByteOpt.Encode16u(temp, dataPackage.KIndex)
	temp = ByteOpt.Encode16u(temp, dataPackage.KCount)
	copy(temp, dataPackage.Data)
	return buffer, nil
}

func UnPackData(dataPackage *DataPackage, buffer []byte) error {
	buffer = ByteOpt.Decode32u(buffer, &dataPackage.Owner)
	buffer = ByteOpt.Decode32u(buffer, &dataPackage.Index)
	buffer = ByteOpt.Decode32u(buffer, &dataPackage.Key)
	buffer = ByteOpt.Decode16u(buffer, &dataPackage.Length)
	buffer = ByteOpt.Decode16u(buffer, &dataPackage.KIndex)
	buffer = ByteOpt.Decode16u(buffer, &dataPackage.KCount)
	dataPackage.Data = buffer
	return nil
}

type ConfirmPackage struct {
	Owner      uint32
	IndexArray UInt32Slice
}

func PackConfirmBytes(confirmPackage *ConfirmPackage) ([]byte, error) {
	buffer := make([]byte, 1+ConfirmPackageHeadSize+2+len(confirmPackage.IndexArray)*4)

	temp := buffer
	temp = ByteOpt.Encode8u(temp, PackType_Confim)
	temp = ByteOpt.Encode32u(temp, confirmPackage.Owner)
	temp = ByteOpt.Encode16u(temp, uint16(len(confirmPackage.IndexArray)))
	for _, value := range confirmPackage.IndexArray {
		temp = ByteOpt.Encode32u(temp, value)
	}
	return buffer, nil
}

func UnPackConfirm(confirmPackage *ConfirmPackage, buffer []byte) error {
	buffer = ByteOpt.Decode32u(buffer, &confirmPackage.Owner)
	var count uint16
	buffer = ByteOpt.Decode16u(buffer, &count)
	confirmPackage.IndexArray = make([]uint32, count)
	for key := range confirmPackage.IndexArray {
		buffer = ByteOpt.Decode32u(buffer, &confirmPackage.IndexArray[key])
	}
	return nil
}

type UInt32Slice []uint32

func (p UInt32Slice) Len() int           { return len(p) }
func (p UInt32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p UInt32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p UInt32Slice) Sort() { sort.Sort(p) }

type ClosePackage struct {
	Owner uint32
}

func PackCloseBytes(closePackage *ClosePackage) ([]byte, error) {
	buffer := make([]byte, 1+4)
	temp := buffer
	temp = ByteOpt.Encode8u(temp, PackType_Close)
	temp = ByteOpt.Encode32u(temp, closePackage.Owner)
	return buffer, nil
}

func UnPackClose(closePackage *ClosePackage, buffer []byte) error {
	buffer = ByteOpt.Decode32u(buffer, &closePackage.Owner)
	return nil
}

type TickPackage struct {
	Owner uint32
	NSec  int64
}

func PackTickPackage(tickPackage *TickPackage) ([]byte, error) {
	buffer := make([]byte, 1+4+8)
	temp := buffer
	temp = ByteOpt.Encode8u(temp, PackType_Tick)
	temp = ByteOpt.Encode32u(temp, tickPackage.Owner)
	temp = ByteOpt.Encode64(temp, tickPackage.NSec)
	return buffer, nil
}

func UnPackTickPackage(tickPackage *TickPackage, buffer []byte) error {
	buffer = ByteOpt.Decode32u(buffer, &tickPackage.Owner)
	buffer = ByteOpt.Decode64(buffer, &tickPackage.NSec)
	return nil
}

type EchoTickPackage struct {
	Owner uint32
	NSec  int64
}

func PackEchoTickPackage(echoTickPackage *EchoTickPackage) ([]byte, error) {
	buffer := make([]byte, 1+4+8)
	temp := buffer
	temp = ByteOpt.Encode8u(temp, PackType_EchoTick)
	temp = ByteOpt.Encode32u(temp, echoTickPackage.Owner)
	temp = ByteOpt.Encode64(temp, echoTickPackage.NSec)
	return buffer, nil
}

func UnPackEchoTickPackage(echoTickPackage *EchoTickPackage, buffer []byte) error {
	buffer = ByteOpt.Decode32u(buffer, &echoTickPackage.Owner)
	buffer = ByteOpt.Decode64(buffer, &echoTickPackage.NSec)
	return nil
}
