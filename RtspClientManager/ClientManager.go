package RtspClientManager

import (

	"net"
	"sync"

)

var ManagerList map[string]*ClientManager = make(map[string]*ClientManager)

type ClientManager struct {
	clients     []*RtspClient
	ClientMutex sync.Mutex
	url         string
}

func NewClientManager(url string) *ClientManager  {
	manager := &ClientManager{url:url}
	ManagerList[url] = manager
	return manager
}

func GetCurrManager(url string) *ClientManager {
	return ManagerList[url]
}

func RemoveManager(url string) {
	ManagerList[url] = nil
}

func (this *ClientManager)DEBUG() {
	/*go func() {
		timer := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timer.C:
				{
					fmt.Println(this.url," client size:", len(this.clients))
				}
			}
		}
	}()*/
}

func (this *ClientManager)Write( data []byte) {
	//ClientMutex.Lock()
	//fmt.Println("client size : ",len(clients))
	for _, value := range this.clients {
		value.Write(data)
	}
	//ClientMutex.Unlock()
}

func (this *ClientManager)AddClient( conn net.Conn) {
	//ClientMutex.Lock()
	c := &RtspClient{start: true, conn: conn}
	this.clients = append(this.clients, c)
	//ClientMutex.Unlock()

}
func (this *ClientManager)RemoveClient( conn net.Conn) {
	//ClientMutex.Lock()

	var list []*RtspClient
	for _, value := range this.clients {
		if value.conn != conn {
			list = append(list, value)
		}

	}
	this.clients = list
	//ClientMutex.Unlock()
}

func (this *ClientManager)GetClients() []*RtspClient {
	return this.clients
}

