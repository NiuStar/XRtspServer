package util

import (
	"net"
	"time"
)

func GetLocalIPAddress() (string, error) {
	conn, err := net.Dial("udp", "www.baidu.com:80")
	if err != nil {
		return "", err
	}

	defer conn.Close()
	return conn.LocalAddr().String(), nil
}

type TimeVal struct {
	Sec  int64
	Usec int64
}

func GetCurrentTimeVal(val *TimeVal) {
	nSec := time.Now().UnixNano()
	val.Sec = nSec / 1000000000
	val.Usec = nSec % (val.Sec * 1000000000)
}
