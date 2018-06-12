package algorithm_test

import (
	"testing"

	"runtime"

	"github.com/sunny352/UDP-go/Utils/SLog"
	"github.com/sunny352/UDP-go/algorithm"
)

func TestAlgoImpl(t *testing.T) {
	mock := algorithm.CreateConnectMock(0, 50, 200)
	//mock.SetDeadline(time.Now().Add(time.Second * 5))
	algo := algorithm.CreateAlgo(mock, 1, 1024)

	msgCount := 200000
	receiveCount := 0
	go func() {
		for receiveCount < msgCount {
			_, err := algo.Receive()
			if nil != err {
				SLog.E(err)
				break
			}
			receiveCount++
		}
	}()

	msg := make([]byte, 4096)
	for key := range msg {
		msg[key] = 'a'
	}
	copy(msg, []byte("ping"))
	for index := 0; index < msgCount; index++ {
		err := algo.Send(msg)
		if nil != err {
			SLog.E(err)
			break
		}
	}

	for algo.Remain() > 0 {
		runtime.Gosched()
	}
}

func BenchmarkAlgoImpl(b *testing.B) {
	mock := algorithm.CreateConnectMock(0, 50, 200)
	//mock.SetDeadline(time.Now().Add(time.Second * 5))
	algo := algorithm.CreateAlgo(mock, 1, 1024)

	msgCount := b.N
	receiveCount := 0
	go func() {
		for receiveCount < msgCount {
			_, err := algo.Receive()
			if nil != err {
				SLog.E(err)
				break
			}
			receiveCount++
		}
	}()

	msg := make([]byte, 1024)
	for key := range msg {
		msg[key] = 'a'
	}
	copy(msg, []byte("ping"))

	SLog.D(b.N)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < msgCount; index++ {
		err := algo.Send(msg)
		if nil != err {
			SLog.E(err)
			break
		}
	}

	b.SetBytes(int64(len(msg)))

	for algo.Remain() > 0 {
		runtime.Gosched()
	}
}
