package rtsp

import (
	"fmt"
	"net"
	"nqc.cn/XRtspServer/RtspClientManager"
	"nqc.cn/XRtspServer/media"
	"nqc.cn/XRtspServer/util"
	"nqc.cn/XRtspServer/sdp"
	"strconv"
	"strings"
	//"time"
)

type RtspClientConnection struct {
	client     *RtspClientManager.RtspClient//这个链接对应的客户端链接
	manager    *RtspClientManager.ClientManager

	control    string
	session    string
	conn       net.Conn
	url        string
	pushClient bool //默认为播放客户端，false，否则为推流客户端
}

func NewRtspClientConnection(conn net.Conn) *RtspClientConnection {
	return &RtspClientConnection{conn: conn, pushClient: false}
}

func (conn *RtspClientConnection) Handle() {
	fmt.Printf("------ rtsp client connection[%s] : handling ------\n", conn.conn.RemoteAddr())

	client := RtspClientManager.NewRtspClient(conn.conn)
	conn.client = client
	client.ReadRequest()
	for {
		select {
		case <-client.Signals:
			fmt.Println("Exit signals by rtsp")
		if !conn.pushClient && conn.manager != nil {
			conn.manager.RemoveClient(conn.conn)
		} else {
			RtspClientManager.RemoveManager(conn.url)
		}
			//RtspClientManager.GetCurrManager(conn.url).RemoveClient(conn.conn)
			fmt.Printf("------ Session[%s] : closed ------\n", conn.conn.RemoteAddr())
			return
		case req := <-client.Outgoing:

			if len(req.URL) != 0 && len(conn.url) == 0 {
				conn.url = req.URL
			}


			resp := conn.handleRequestAndReturnResponse(req)
			if resp != nil {
				//time.Sleep(1 * time.Second)
				_, err := conn.conn.Write([]byte(resp.String()))
				if err != nil && !conn.pushClient && conn.manager != nil {
					conn.manager.RemoveClient(conn.conn)
					conn.conn.Close()
					return
				} else if err != nil {
					conn.conn.Close()
					return
				}
				fmt.Printf("------ rtsp client connection[%s] : get request ------ \n%s\n", conn.conn.RemoteAddr(), req)
				fmt.Printf("------ Session[%s] : set response ------ \n%s\n", conn.conn.RemoteAddr(), resp)
			}
			//处理RTSP请求
			if req.Method != RtspClientManager.DATA {

				if req.Method != RtspClientManager.PLAY &&  req.Method != RtspClientManager.RECORD {
					fmt.Println("Player ")
					client.ReadRequest()
				}
			}
		}
	}
}

func (conn *RtspClientConnection) handleRequestAndReturnResponse(req *RtspClientManager.Request) *RtspClientManager.Response {

	cSeq := req.Header.Get("CSeq")


	switch req.Method {
	case RtspClientManager.DATA:
		if   conn.manager != nil {
		conn.manager.Write([]byte(req.Body))
	}
		//RtspClientManager.Write( []byte(req.Body))
		return nil
	case RtspClientManager.ANNOUNCE:
		{
			fmt.Println(",req.Content : ", req.Content)
			//
			infos := sdp.Decode(req.Content)
			for _,info := range infos {
				if strings.EqualFold(info.AVType,"video") {
					conn.control = info.Control
				}
			}

			sdpName := util.GetSdpName(req.URL)

			_, exits := media.NewMediaSession(sdpName, req.Body)
			if exits != nil {
				fmt.Println(exits)
			}
			manager := RtspClientManager.NewClientManager(sdpName)
			conn.manager = manager
			conn.manager.DEBUG()
			conn.url = req.URL

			if len(cSeq) == 0 {
				cSeq = "0"
			}
			//seq,_ := strconv.ParseInt(cSeq,10,64)
			//cSeq = strconv.FormatInt(seq + 1,10)


			resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
			if resp != nil {
				//time.Sleep(1 * time.Second)
				//conn.conn.Write([]byte(resp.String()))
			}

			//fmt.Printf("------ Session[%s] : set response ------ \n%s\n", conn.conn.RemoteAddr(), resp)
		}
		fmt.Printf("------conn.pushClient = true---------- ")
		conn.pushClient = true

	case RtspClientManager.OPTIONS:
		resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
		options := strings.Join([]string{RtspClientManager.OPTIONS, RtspClientManager.DESCRIBE, RtspClientManager.SETUP, RtspClientManager.PLAY, RtspClientManager.TEARDOWN,RtspClientManager.RECORD}, ", ")
		resp.Header["Public"] = []string{options}
		return resp

	case RtspClientManager.DESCRIBE:
		sdpName := util.GetSdpName(req.URL)
		mediaSess, exits := media.LookupMediaSession(sdpName)
		if !exits {
			return RtspClientManager.NewResponse(RtspClientManager.SessionNotFound, "Session not found", cSeq, "")
		}
		sdp := mediaSess.GenerateSDPDescription()

		resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
		resp.Header.Add("Content-Base", req.URL)
		resp.Header.Add("Content-Type", "application/sdp")
		resp.Header.Add("Content-Length", strconv.Itoa(len(sdp)))
		resp.Body = sdp
		return resp

	case RtspClientManager.SETUP:


		//fmt.Println("接收到 cSeq:" + cSeq)

		resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
		resp.Header.Add("Transport", req.Header.Get("Transport"))
		Session := req.Header.Get("Session")
		if len(Session) <= 0 {
			Session = "1"
		}
		resp.Header.Add("Session",Session)
		resp.Header.Add("Server", "XVideoStreamServer")
		resp.Header.Add("Cache-Control", "no-cache")
		conn.session = Session
		//fmt.Println("返回的 cSeq:" + resp.String())
		return resp
	case RtspClientManager.RECORD:
		{
			conn.client.PushLayer()
			fmt.Printf("------conn.client.PushLayer()---------- ")
			resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
			if resp != nil {
				//conn.conn.Write([]byte(resp.String()))
			}
			resp.Header.Add("Session", conn.session)
			resp.Header.Add("RTP-Info", "url=" + conn.url + "/" + conn.control)
			go conn.client.ReadData()
			return resp
		}
	case RtspClientManager.PLAY:

		//time.Sleep(2 * time.Second)
		if conn.pushClient {
			conn.client.PushLayer()
			fmt.Printf("------conn.client.PushLayer()---------- ")
			resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
			if resp != nil {

			}
			go conn.client.ReadData()
			return resp
		} else {
			fmt.Printf("------  !conn.client.PushLayer()---------- ")
			sdpName := util.GetSdpName(req.URL)
			conn.manager = RtspClientManager.GetCurrManager(sdpName)
			if  conn.manager != nil {
				conn.manager.AddClient( conn.conn)
			}
			//go conn.client.ReadData()

		}

		break
	case RtspClientManager.TEARDOWN:
		fmt.Println("TEARDOWN")
		if !conn.pushClient  && conn.manager != nil {
			conn.manager.RemoveClient(conn.conn)
		}
		//conn.manager.RemoveClient(conn.url, conn.conn)
		break

	/*case "":
		{
			return RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
		}*/
	default:
		return RtspClientManager.NewResponse(RtspClientManager.MethodNotAllowed, "Option Not Support", cSeq, "")
	}

	return RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
}
