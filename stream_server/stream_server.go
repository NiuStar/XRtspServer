package stream_server

import (
	"github.com/NiuStar/XRtspServer/rtsp"
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

func (s *StreamServer) NewRtspServer() (*rtsp.RtspServer, error) {
	return rtsp.NewRtspServer(s.addr)
}

func (s *StreamServer) Run(rtspServer *rtsp.RtspServer) {

	s.Wrap(func() {
		rtspServer.Run()
	})

	s.wrap.Wait()
}
