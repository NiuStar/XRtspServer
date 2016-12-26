package rtsp

import (
	"fmt"
	"net"
	"runtime"

)

type RtspServer struct {
	tcpListener net.Listener
}

func NewRtspServer(address string) (*RtspServer, error) {
	server := &RtspServer{}

	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("ERROR: listen (", address, ") failed -", err)
		return nil, err
	}

	server.tcpListener = tcpListener

	return server, nil
}

func (s *RtspServer) Run() {
	fmt.Println("RTSP Listen on", s.tcpListener.Addr())

	for {
		clientConn, err := s.tcpListener.Accept()
		if err != nil {
			//若是暂时性错误，则继续监听，否则直接退出
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				fmt.Println("NOTICE: temporary Accept() failure -", err)
				runtime.Gosched()
				continue
			}

			break
		}

		conn := NewRtspClientConnection(clientConn)

		go conn.Handle()
	}

	fmt.Println("RTSP Stop listenning on", s.tcpListener.Addr())
}
