package rtsp

import (
	"fmt"
	"net"
	"nqc.cn/XRtspServer/RtspClientManager"
	"nqc.cn/XRtspServer/media"
	"strconv"
	"strings"
)

type RtspClientConnection struct {
	client     *RtspClientManager.RtspClient
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
	go client.ReadRequest()
	for {
		select {
		case <-client.Signals:
			fmt.Println("Exit signals by rtsp")
			RtspClientManager.RemoveClient(conn.url, conn.conn)
			fmt.Printf("------ Session[%s] : closed ------\n", conn.conn.RemoteAddr())
			return
		case req := <-client.Outgoing:

			if len(req.URL) != 0 && len(conn.url) == 0 {
				conn.url = req.URL
			}

			//处理RTSP请求
			if req.Method != RtspClientManager.DATA {
				fmt.Printf("------ rtsp client connection[%s] : get request ------ \n%s\n", conn.conn.RemoteAddr(), req)
			}

			resp := conn.handleRequestAndReturnResponse(req)
			if resp != nil {

				_, err := conn.conn.Write([]byte(resp.String()))
				if err != nil {
					RtspClientManager.RemoveClient(conn.url, conn.conn)
					conn.conn.Close()
				}
				fmt.Printf("------ Session[%s] : set response ------ \n%s\n", conn.conn.RemoteAddr(), resp)
			}

		}
	}
	/*for {
		req, err := ReadRequest(conn.conn)
		if err != nil {
			break
		}
		if len(req.URL) != 0 && len(conn.url) == 0 {
			conn.url = req.URL
		}
		fmt.Printf("------ rtsp client connection[%s] : get request ------ \n%s\n", conn.conn.RemoteAddr(), req)
		//处理RTSP请求
		resp := conn.handleRequestAndReturnResponse(req)
		if resp != nil {
			conn.conn.Write([]byte(resp.String()))
		} else {
			break
		}

		fmt.Printf("------ Session[%s] : set response ------ \n%s\n", conn.conn.RemoteAddr(), resp)
	}*/

	//fmt.Printf("------ Session[%s] : closed ------\n", conn.conn.RemoteAddr())
}

func (conn *RtspClientConnection) handleRequestAndReturnResponse(req *RtspClientManager.Request) *RtspClientManager.Response {

	cSeq := req.Header.Get("CSeq")

	switch req.Method {
	case RtspClientManager.DATA:
		RtspClientManager.Write(conn.url, []byte(req.Body))
		return nil
	case RtspClientManager.ANNOUNCE:
		{
			fmt.Println(",req.Body : ", req.Body)
			_, exits := media.NewMediaSession(req.URL, req.Body)
			if exits != nil {
				fmt.Println(exits)
			}
			RtspClientManager.DEBUG(req.URL)
			conn.url = req.URL
			resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
			if resp != nil {
				conn.conn.Write([]byte(resp.String()))
			}

			fmt.Printf("------ Session[%s] : set response ------ \n%s\n", conn.conn.RemoteAddr(), resp)
		}
		conn.pushClient = true

	case RtspClientManager.OPTIONS:
		resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
		options := strings.Join([]string{RtspClientManager.OPTIONS, RtspClientManager.DESCRIBE, RtspClientManager.SETUP, RtspClientManager.PLAY, RtspClientManager.TEARDOWN}, ", ")
		resp.Header["Public"] = []string{options}
		return resp

	case RtspClientManager.DESCRIBE:

		mediaSess, exits := media.LookupMediaSession(req.URL)
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

		resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
		resp.Header.Add("Transport", req.Header.Get("Transport"))
		Session := req.Header.Get("Session")
		if len(Session) > 0 {
			resp.Header.Add("Session", req.Header.Get("Session"))
		} else {
			resp.Header.Add("Session", "1")
		}
		resp.Header.Add("Server", "XVideoStreamServer")
		resp.Header.Add("Cache-Control", "no-cache")
		return resp
	case RtspClientManager.PLAY:
		if conn.pushClient {
			conn.client.PushLayer()
			resp := RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
			if resp != nil {
				conn.conn.Write([]byte(resp.String()))
			}
			return resp
		} else {
			RtspClientManager.AddClient(conn.url, conn.conn)
		}
		break
	case RtspClientManager.TEARDOWN:
		fmt.Println("TEARDOWN")
		RtspClientManager.RemoveClient(conn.url, conn.conn)
		break
	default:
		return RtspClientManager.NewResponse(RtspClientManager.MethodNotAllowed, "Option Not Support", cSeq, "")
	}

	return RtspClientManager.NewResponse(RtspClientManager.OK, "OK", cSeq, "")
}
