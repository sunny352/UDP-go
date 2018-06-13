package algorithm_test

import (
	"testing"

	"runtime"

	"fmt"

	"github.com/sunny352/UDP-go/Utils/SLog"
	"github.com/sunny352/UDP-go/Utils/Safe"
	"github.com/sunny352/UDP-go/algorithm"
	"github.com/sunny352/UDP-go/network"
)

func TestAlgoImpl(t *testing.T) {
	mock := algorithm.CreateConnectMock(0, 50, 200)
	//mock.SetDeadline(time.Now().Add(time.Second * 5))
	algo := algorithm.CreateAlgo(mock, 1, 1024)

	msgCount := 1024
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

	msg := make([]byte, 4096)
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

func TestAlgoNetwork(t *testing.T) {
	listener, err := network.Listen("udp", ":0")
	if nil != err {
		t.Error(err)
		return
	}

	Safe.Go(func() {
		for {
			connect, err := listener.Accept()
			if nil != err {
				t.Error(err)
				break
			}
			receive := algorithm.CreateAlgo(connect, 1, 1024)
			count := 0
			for {
				msg, err := receive.Receive()
				if nil != err {
					t.Error(err)
					break
				}
				count++
				SLog.D(string(msg), count)
			}
		}
	})

	sendConnect, err := network.Dial("udp", listener.Addr().String())
	if nil != err {
		t.Error(err)
	}

	algo := algorithm.CreateAlgo(sendConnect, 1, 1024)
	//msg := make([]byte, 1024)

	for index := 0; index < 10240; index++ {
		algo.Send([]byte(fmt.Sprintf("msg_%d", index)))
	}
	//SLog.D(algo.Remain())
	for algo.Remain() > 0 {
		//SLog.D(algo.Remain())
		runtime.Gosched()
	}
}

func BenchmarkAlgoNetwork(b *testing.B) {
	listener, err := network.Listen("udp", ":0")
	if nil != err {
		b.Error(err)
	}

	Safe.Go(func() {
		for {
			connect, err := listener.Accept()
			if nil != err {
				b.Error(err)
				break
			}
			//algorithm.CreateAlgo(connect, 1, 1024)
			receive := algorithm.CreateAlgo(connect, 1, 1024)
			count := 0
			for {
				_, err := receive.Receive()
				if nil != err {
					b.Error(err)
					break
				}
				count++
			}
		}
	})

	sendConnect, err := network.Dial("udp", listener.Addr().String())
	if nil != err {
		b.Error(err)
	}

	algo := algorithm.CreateAlgo(sendConnect, 1, 1024)
	msg := make([]byte, 1024)

	SLog.D(b.N)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		algo.Send(msg)
	}
	b.SetBytes(int64(len(msg)))
	for algo.Remain() > 0 {
		runtime.Gosched()
	}
}
