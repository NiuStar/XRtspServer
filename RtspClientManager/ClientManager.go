package RtspClientManager

import (
	"sync"
	"time"
	"net"
	"fmt"
)

var (
	 clients map[string][]*RtspClient = make(map[string][]*RtspClient)
	 ClientMutex   sync.Mutex
)

func DEBUG(url string) {
	go func() {
		timer := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timer.C:
				{
					fmt.Println("client size:",len(clients[url]))
				}
			}
		}
	}()
}

func Write(url string,data []byte) {
	//ClientMutex.Lock()
	//fmt.Println("client size : ",len(clients))
	for _,value := range clients[url] {
		value.Write(data)
	}
	//ClientMutex.Unlock()
}

func AddClient(url string,conn net.Conn) {
	//ClientMutex.Lock()
	c := &RtspClient{start:true,conn:conn}
	clients[url] = append(clients[url],c)
	//ClientMutex.Unlock()

}
func RemoveClient(url string,conn net.Conn) {
	//ClientMutex.Lock()

	var list []*RtspClient
	for _,value := range clients[url] {
		if value.conn != conn {
			list = append(list,value)
		}

	}
	clients[url] = list
	//ClientMutex.Unlock()

}
