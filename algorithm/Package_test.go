package algorithm_test

import (
	"math"
	"testing"

	"github.com/sunny352/UDP-go/Utils/SLog"
	"github.com/sunny352/UDP-go/algorithm"
)

func TestSeparatePackage(t *testing.T) {
	data := make([]byte, 1024)
	for index := 0; index < len(data); index++ {
		data[index] = byte(index % math.MaxInt8)
	}
	SLog.D(data)
	packList, _, err := algorithm.SeparatePackage(data, 1, 100, 5, 1024)
	if nil != err {
		t.Error(err)
	}
	for _, value := range packList {
		SLog.D(value.Key, value.Index, value.KIndex, value.KCount, value.Length, value.Data)
	}

	first := packList[0]

	combiner := algorithm.Combiner{}
	combiner.Init(first.KCount)
	for _, value := range packList {
		if combiner.Push(&value) {
			combinerBytes := combiner.Bytes()
			if len(combinerBytes) != len(data) {
				t.Error("长度不同")
				return
			}
			SLog.D("Success")
			SLog.D(combinerBytes)
			return
		}
	}
	t.Error("合并错误")
}

//func BenchmarkSeparatePackage(b *testing.B) {
//	data := make([]byte, 10240)
//	for index := 0; index < len(data); index++ {
//		data[index] = byte(index % math.MaxInt8)
//	}
//	b.ReportAllocs()
//	for index := 0; index < b.N; index++ {
//		algorithm.SeparatePackage(data, 1, 100, 5, 1024)
//	}
//}
//
//func BenchmarkCombiner(b *testing.B) {
//	data := make([]byte, 10240)
//	for index := 0; index < len(data); index++ {
//		data[index] = byte(index % math.MaxInt8)
//	}
//	packList, _ := algorithm.SeparatePackage(data, 1, 100, 5, 1024)
//	b.ReportAllocs()
//	for index := 0; index < b.N; index++ {
//		combiner := algorithm.Combiner{}
//		combiner.Init(packList[0].KCount)
//		for _, value := range packList {
//			combiner.Push(&value)
//		}
//		combiner.Bytes()
//	}
//}

func BenchmarkSeparateCombiner_64k_512(b *testing.B) {
	benchSepComb(b, 65536, 512)
}
func BenchmarkSeparateCombiner_64k_1024(b *testing.B) {
	benchSepComb(b, 65536, 1024)
}
func BenchmarkSeparateCombiner_64k_256(b *testing.B) {
	benchSepComb(b, 65536, 256)
}
func BenchmarkSeparateCombiner_1M_512(b *testing.B) {
	benchSepComb(b, 1048576, 512)
}
func BenchmarkSeparateCombiner_1M_1024(b *testing.B) {
	benchSepComb(b, 1048576, 1024)
}
func BenchmarkSeparateCombiner_1M_256(b *testing.B) {
	benchSepComb(b, 1048576, 256)
}
func BenchmarkSeparateCombiner_256k_1024(b *testing.B) {
	benchSepComb(b, 524288, 1024)
}
func BenchmarkSeparateCombiner_1k_1024(b *testing.B) {
	benchSepComb(b, 1024, 1024)
}
func BenchmarkSeparateCombiner_16k_1024(b *testing.B) {
	benchSepComb(b, 1024*16, 1024)
}
func BenchmarkSeparateCombiner_128k_1024(b *testing.B) {
	benchSepComb(b, 1024*128, 1024)
}

func benchSepComb(b *testing.B, dataSize int, sepLength uint16) {
	data := make([]byte, dataSize)
	for index := 0; index < len(data); index++ {
		data[index] = byte(index % math.MaxInt8)
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		packList, _, _ := algorithm.SeparatePackage(data, 1, 100, 5, sepLength)
		combiner := algorithm.Combiner{}
		combiner.Init(packList[0].KCount)
		for _, value := range packList {
			combiner.Push(&value)
		}
		combiner.Bytes()
	}
	b.SetBytes(int64(len(data)))
}

func TestPackage(t *testing.T) {
	var dataPackage algorithm.DataPackage
	dataPackage.Owner = 1
	packBytes, err := algorithm.PackDataBytes(&dataPackage)
	if nil != err {
		t.Error(err)
	}
	if nil == packBytes {
		t.Error("fail")
	}
	SLog.D(packBytes)
}

func TestPackDataBytes(t *testing.T) {
	dataPackage := algorithm.DataPackage{
		Owner:  1,
		Index:  100,
		Key:    10,
		KIndex: 0,
		KCount: 1,
		Length: 512,
		Data:   make([]byte, 512),
	}
	dataBytes, _ := algorithm.PackDataBytes(&dataPackage)
	SLog.D(dataBytes)
	newData := algorithm.DataPackage{}
	algorithm.UnPackData(&newData, dataBytes[1:])
	SLog.D(newData.Owner)
}
