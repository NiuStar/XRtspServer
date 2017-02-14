package RtspClientManager

import (
	"net"
	"fmt"
)

type RtspClient struct {
	start bool
	conn net.Conn

	pushClient bool //默认为播放客户端，false，否则为推流客户端
	Signals       chan bool
	Outgoing      chan *Request
}

func NewRtspClient(conn net.Conn) *RtspClient {
	return &RtspClient{conn: conn, start: true,pushClient:false, Signals: make(chan bool, 1), Outgoing: make(chan *Request, 1)}
}

func (c *RtspClient)Write(data[] byte) {
	if c.start {
		if data[0] == 36 && data[1] == 0 {
			cc := data[4] & 0xF
			//rtp header
			rtphdr := 12 + cc*4

			nalType := data[4+rtphdr] & 0x1F
			if nalType == 28 {
				isStart := data[4+rtphdr+1]&0x80 != 0

				if isStart {
					c.start = false
				} else {
					return
				}
			}
		}
	}
	c.conn.Write(data)
}

func (conn *RtspClient) PushLayer() {
	conn.pushClient = true
}


func (this *RtspClient) ReadRequest() {


	if !this.pushClient {

		req, err := ReadRequest(this.conn)
		if err != nil {
			//panic(err)
			fmt.Println("err")
			return
		}
		fmt.Println("no err",req)

		this.Outgoing <- req
	} else {
		data,_ := ReadSocket(this.conn)

		if data != nil {
			req := &Request{
				Header: make(map[string][]string),
			}
			req.Method = DATA
			req.Body = string(data)
			//fmt.Println(data)
			this.Outgoing <- req
		} else {
			return
		}
	}

}

func (this *RtspClient) ReadData() {
	defer func() {
		this.Signals <- true
	}()

	for {
		/*if !this.pushClient {
			req, err := ReadRequest(this.conn)
			if err != nil {
				//panic(err)
				fmt.Println("err")
				break
			}
			fmt.Println("no err")
			this.Outgoing <- req
		} else */{
			data,err := ReadSocket(this.conn)
			if err != nil {
				return
			}
			if data != nil {
				req := &Request{
					Header: make(map[string][]string),
				}
				req.Method = DATA
				req.Body = string(data)
				//fmt.Println(data)
				this.Outgoing <- req
			} else {
				continue
			}
		}
	}
}