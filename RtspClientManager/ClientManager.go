package RtspClientManager

import (
	"net"
	"sync"

	"github.com/NiuStar/log/fmt"
)

var ManagerList map[string]*ClientManager = make(map[string]*ClientManager)
var ManagerMutex sync.Mutex

type ClientManager struct {
	clients     []*RtspClient
	ClientS     []string
	clientMutex sync.Mutex
	Url         string
}

func NewClientManager(url string) *ClientManager {
	manager := &ClientManager{Url: url}
	ManagerMutex.Lock()
	ManagerList[url] = manager
	ManagerMutex.Unlock()

	fmt.Println("ManagerList1:", ManagerList)
	return manager
}

func GetCurrManager(url string) *ClientManager {
	ManagerMutex.Lock()
	list := ManagerList[url]
	ManagerMutex.Unlock()
	//fmt.Println("ManagerList3:",ManagerList)
	return list
}

func GetCurrManagers() []*ClientManager {
	var list []*ClientManager
	ManagerMutex.Lock()
	for _, value := range ManagerList {
		list = append(list, value)
	}
	ManagerMutex.Unlock()
	fmt.Println("ManagerList3:", ManagerList)
	return list
}

func RemoveManager(url string) {
	ManagerMutex.Lock()
	delete(ManagerList, url)
	fmt.Println("ManagerList2:", ManagerList)
	ManagerMutex.Unlock()
}

func (this *ClientManager) DEBUG() {
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

func (this *ClientManager) Write(data []byte) {
	//ClientMutex.Lock()
	//fmt.Println("client size : ",len(clients))
	for _, value := range this.clients {
		value.Write(data)
	}
	//ClientMutex.Unlock()
}

func (this *ClientManager) AddClient(conn net.Conn) {

	fmt.Println("AddClient。。。。。。。。。")
	//ClientMutex.Lock()
	c := &RtspClient{start: true, conn: conn}

	this.clients = append(this.clients, c)
	this.ClientS = append(this.ClientS, conn.RemoteAddr().String())
	//ClientMutex.Unlock()

}
func (this *ClientManager) RemoveClient(conn net.Conn) {
	//ClientMutex.Lock()

	var list []*RtspClient
	var cl []string
	for _, value := range this.clients {
		if value.conn != conn {
			list = append(list, value)
			cl = append(cl, value.conn.RemoteAddr().String())
		}

	}
	this.ClientS = cl
	this.clients = list
	//ClientMutex.Unlock()
}

func (this *ClientManager) GetClients() []*RtspClient {
	return this.clients
}
