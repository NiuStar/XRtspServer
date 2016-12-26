package stream_server

import (
	"nqc.cn/XRtspServer/rtsp"
	"sync"
)

type StreamServer struct {
	addr string
	wrap sync.WaitGroup
}

func NewStreamServer(addr string) *StreamServer {
	server := &StreamServer{addr: addr}
	return server
}

func (s *StreamServer) Wrap(cb func()) {
	s.wrap.Add(1)
	go func() {
		cb()
		s.wrap.Done()
	}()
}

func (s *StreamServer) Run() {
	rtspServer, err := rtsp.NewRtspServer(s.addr)
	if err != nil {
		return
	}

	s.Wrap(func() {
		rtspServer.Run()
	})

	s.wrap.Wait()
}
