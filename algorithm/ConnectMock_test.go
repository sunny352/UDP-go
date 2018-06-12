package algorithm_test

import (
	"testing"

	"github.com/sunny352/UDP-go/Utils/SLog"
	"github.com/sunny352/UDP-go/algorithm"
)

func TestMock(t *testing.T) {
	connect := algorithm.CreateConnectMock(0, 0, 0)
	defer connect.Close()

	msgCount := 65535

	temp := make([]byte, 1024)
	wait := make(chan interface{})
	go func() {
		defer func() {
			wait <- 1
		}()
		count := 0
		for count < msgCount {
			_, err := connect.Read(temp)
			if nil != err {
				SLog.E(err)
				break
			}
			count++
		}
	}()

	msg := make([]byte, 1024)
	for index := 0; index < msgCount; index++ {
		connect.Write(msg)
	}

	<-wait
}

func BenchmarkConnectMock(b *testing.B) {
	connect := algorithm.CreateConnectMock(0, 0, 0)
	defer connect.Close()

	msgCount := b.N
	msgLength := 1024

	temp := make([]byte, msgLength)
	wait := make(chan interface{})
	go func() {
		defer func() {
			wait <- 1
		}()
		count := 0
		for count < msgCount {
			_, err := connect.Read(temp)
			if nil != err {
				SLog.E(err)
				break
			}
			count++
		}
	}()

	msg := make([]byte, msgLength)

	SLog.D(b.N)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < msgCount; index++ {
		connect.Write(msg)
	}

	b.SetBytes(int64(len(msg)))

	<-wait
}
