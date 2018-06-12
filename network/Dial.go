package network

import "net"

func Dial(network, address string) (Connect, error) {
	addr, err := net.ResolveUDPAddr(network, address)
	if nil != err {
		return nil, err
	}
	udp, err := net.DialUDP(network, nil, addr)
	if nil != err {
		return nil, err
	}
	//udp.SetReadBuffer(1024 * 1024)
	//udp.SetWriteBuffer(1024 * 1024)
	return CreateConnectWithoutListener(addr, udp)
}
