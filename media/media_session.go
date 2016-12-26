package media

import (
	//"fmt"
	"sync"

	"github.com/stream_server/util"
)

type MediaSession struct {
	streamName string
	timeval    util.TimeVal
	ipAddress  string

	mdata       string
}

var (
	MapMutex   sync.Mutex
	SessionMap map[string]*MediaSession
)

func init() {
	SessionMap = make(map[string]*MediaSession)
}

func NewMediaSession(streamName string,data string) (*MediaSession, error) {


	sess := &MediaSession{
		streamName: streamName,
		mdata:data,
	}

	MapMutex.Lock()
	SessionMap[streamName] = sess
	MapMutex.Unlock()

	return sess, nil
}

func LookupMediaSession(streamName string) (*MediaSession, bool) {
	MapMutex.Lock()
	v, ok := SessionMap[streamName]
	MapMutex.Unlock()
	return v, ok
}

func (sess *MediaSession) GenerateSDPDescription() string {
	return sess.mdata
}

