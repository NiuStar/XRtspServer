package RtspClientManager

import (
	//"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	RTSP_VERSION = "RTSP/1.0"
)

const (
	// Client to server for presentation and stream objects; recommended
	DESCRIBE = "DESCRIBE"
	// Bidirectional for client and stream objects; optional
	ANNOUNCE = "ANNOUNCE"
	// Bidirectional for client and stream objects; optional
	GET_PARAMETER = "GET_PARAMETER"
	// Bidirectional for client and stream objects; required for Client to server, optional for server to client
	OPTIONS = "OPTIONS"
	// Client to server for presentation and stream objects; recommended
	PAUSE = "PAUSE"
	// Client to server for presentation and stream objects; required
	PLAY = "PLAY"
	// Client to server for presentation and stream objects; optional
	RECORD = "RECORD"
	// Server to client for presentation and stream objects; optional
	REDIRECT = "REDIRECT"
	// Client to server for stream objects; required
	SETUP = "SETUP"
	// Bidirectional for presentation and stream objects; optional
	SET_PARAMETER = "SET_PARAMETER"
	// Client to server for presentation and stream objects; required
	TEARDOWN = "TEARDOWN"
	DATA = "DATA"
)

const ()

const (
	// all requests
	Continue = 100

	// all requests
	OK = 200
	// RECORD
	Created = 201
	// RECORD
	LowOnStorageSpace = 250

	// all requests
	MultipleChoices = 300
	// all requests
	MovedPermanently = 301
	// all requests
	MovedTemporarily = 302
	// all requests
	SeeOther = 303
	// all requests
	UseProxy = 305

	// all requests
	BadRequest = 400
	// all requests
	Unauthorized = 401
	// all requests
	PaymentRequired = 402
	// all requests
	Forbidden = 403
	// all requests
	NotFound = 404
	// all requests
	MethodNotAllowed = 405
	// all requests
	NotAcceptable = 406
	// all requests
	ProxyAuthenticationRequired = 407
	// all requests
	RequestTimeout = 408
	// all requests
	Gone = 410
	// all requests
	LengthRequired = 411
	// DESCRIBE, SETUP
	PreconditionFailed = 412
	// all requests
	RequestEntityTooLarge = 413
	// all requests
	RequestURITooLong = 414
	// all requests
	UnsupportedMediaType = 415
	// SETUP
	Invalidparameter = 451
	// SETUP
	IllegalConferenceIdentifier = 452
	// SETUP
	NotEnoughBandwidth = 453
	// all requests
	SessionNotFound = 454
	// all requests
	MethodNotValidInThisState = 455
	// all requests
	HeaderFieldNotValid = 456
	// PLAY
	InvalidRange = 457
	// SET_PARAMETER
	ParameterIsReadOnly = 458
	// all requests
	AggregateOperationNotAllowed = 459
	// all requests
	OnlyAggregateOperationAllowed = 460
	// all requests
	UnsupportedTransport = 461
	// all requests
	DestinationUnreachable = 462

	// all requests
	InternalServerError = 500
	// all requests
	NotImplemented = 501
	// all requests
	BadGateway = 502
	// all requests
	ServiceUnavailable = 503
	// all requests
	GatewayTimeout = 504
	// all requests
	RTSPVersionNotSupported = 505
	// all requests
	OptionNotsupport = 551
)

// RTSP请求的格式
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 方法  | <空格>  | URL | <空格>  | 版本  | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值     | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                   ......                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值     | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | <回车换行>                                            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  实体内容                                             |
// |  （通常不用）                                         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type Request struct {
	Method  string
	URL     string
	Version string
	Header  http.Header
	Body    string
}

func NewRequest(method, url, cSeq, body string) *Request {
	req := &Request{
		Method:  method,
		URL:     url,
		Version: RTSP_VERSION,
		Header:  map[string][]string{"CSeq": []string{cSeq}},
		Body:    body,
	}
	return req
}

func (r *Request) String() string {
	str := fmt.Sprintf("%s %s %s\r\n", r.Method, r.URL, r.Version)
	for key, values := range r.Header {
		for _, value := range values {
			str += fmt.Sprintf("%s: %s\r\n", key, value)
		}
	}
	str += "\r\n"
	str += r.Body
	return str
}

func getAllSocket(r io.Reader) []byte {
	header := make([]byte, 4)
	payload := make([]byte, 16384)
	sync_b := make([]byte, 1)
	//timer := time.Now()


	for {
		if n, err := io.ReadFull(r, header); err != nil || n != 4 {

				fmt.Println("read header error", err)

			return nil
		}
		if header[0] != 36 {
			fmt.Println("header[0] != 36")
			rtsp := false
			if string(header) != "RTSP" {

					fmt.Println("desync strange data repair", string(header), header)
				return nil

			} else {
				rtsp = true
			}
			i := 1
			for {
				i++
				if i > 4096 {

					fmt.Println("desync fatal miss position rtp packet")

					return nil
				}
				if n, err := io.ReadFull(r, sync_b); err != nil && n != 1 {
					return nil
				}
				if sync_b[0] == 36 {
					header[0] = sync_b[0]
					if n, err := io.ReadFull(r, sync_b); err != nil && n != 1 {
						return nil
					}
					if sync_b[0] == 0 || sync_b[0] == 1 || sync_b[0] == 2 || sync_b[0] == 3 {
						header[1] = sync_b[0]
						if n, err := io.ReadFull(r, header[2:]); err != nil && n == 2 {
							return nil
						}
						if !rtsp {

								fmt.Println("desync fixed ok", sync_b[0], i, "afrer byte")

						}
						break
					} else {

						fmt.Println("desync repair fail chanel incorect", sync_b[0])

					}
				}
			}
			//fmt.Println("rtsp : ",rtsp)
		} else {
			//	fmt.Println("header[0] == 36:",header[0],header[1],header[2],header[3])
			//fmt.Println("header[0] == 36")
		}


		payloadLen := (int)(header[2])<<8 + (int)(header[3])
		if payloadLen > 16384 || payloadLen < 12 {

				fmt.Println("fatal size desync",  payloadLen)

			continue
		}
		if n, err := io.ReadFull(r, payload[:payloadLen]); err != nil || n != payloadLen {

				fmt.Println("read payload error", payloadLen, err)

			return nil
		} else {

			return append(header, payload[:n]...)
		}
	}
}

func ReadSocket(r io.Reader)(data []byte) {

	return getAllSocket(r)


}

func ReadRequest(r io.Reader) (req *Request, err error) {

	req = &Request{
		Header: make(map[string][]string),
	}

	/*
	buffer := make([]byte, 2048)
	len, err := r.Read(buffer)
	if err != nil && len <= 0 {
		return nil, err
	}*/
/*	buffer := getAllSocket(r)

	if buffer == nil {
		buffer_i := make([]byte, 2048)
		len, err := r.Read(buffer_i)
		if err != nil && len <= 0 {
			return nil, err
		}
		buffer = buffer_i[:len]
	}

	if buffer[0] == 36 && buffer[1] == 0  {
		req.Method = "DATA"
		req.Body = string(buffer)
		return req, nil

	} else if buffer[0] == 36 && buffer[1] == 1 {
		req.Method = "DATA"
		req.Body = string(buffer)
		return req, nil
	}*/
	buffer_i := make([]byte, 2048)
	leni, err := r.Read(buffer_i)
	if err != nil && leni <= 0 {
		return nil, err
	}
	buffer := buffer_i[:leni]

	recv := string(buffer)
//	fmt.Println("recv:",recv)

	context := strings.SplitN(recv, "\r\n\r\n", 2)
	header := context[0]
	var body string
	if len(context) > 1 {
		body = context[1]
	}


	parts := strings.SplitN(header, "\r\n", 2)
	dest := parts[0]
	var prop string
	if len(parts) > 1 {
		prop = parts[1]
	}

	parts = strings.SplitN(dest, " ", 3)
	req.Method = parts[0]
	if len(parts) > 2 {
		req.URL = parts[1]
		req.Version = parts[2]
	}

	pairs := strings.Split(prop, "\r\n")
	for _, pair := range pairs {
		parts = strings.SplitN(pair, ": ", 2)
		key := parts[0]
		if len(parts) > 1 {
			value := parts[1]
			req.Header.Add(key, value)
		} else {
			req.Header.Add(key, "")
		}


	}

	req.Body = string(body)

	return req, nil
}

// RTSP响应的格式
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 版本  | <空格>  | 状态码 | <空格>  | 状态描述  | <回车换行> |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值           | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           ......                            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | 头部字段名  | : | <空格>  |      值           | <回车换行>  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | <回车换行>                                                  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  实体内容                                                   |
// |  （有些响应不用）                                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type Response struct {
	Version    string
	StatusCode int
	Status     string
	Header     http.Header
	Body       string
}

func NewResponse(statusCode int, status, cSeq, body string) *Response {
	res := &Response{
		Version:    RTSP_VERSION,
		StatusCode: statusCode,
		Status:     status,
		Header:     map[string][]string{"CSeq": []string{cSeq}},
		Body:       body,
	}
	return res
}

func (r *Response) String() string {
	str := fmt.Sprintf("%s %d %s\r\n", r.Version, r.StatusCode, r.Status)
	for key, values := range r.Header {
		for _, value := range values {
			str += fmt.Sprintf("%s: %s\r\n", key, value)
		}
	}
	str += "\r\n"
	str += r.Body
	return str
}

func ReadResponse(r io.Reader) (res *Response, err error) {


	res = &Response{
		Header: make(map[string][]string),
	}

	buffer := make([]byte, 2048)
	len, err := r.Read(buffer)
	if err != nil && len <= 0 {
		return nil, err
	}
	recv := string(buffer[:len])


	fmt.Println("recv:",recv)
	context := strings.SplitN(recv, "\r\n\r\n", 2)
	header := context[0]
	body := context[1]

	parts := strings.SplitN(header, "\r\n", 2)
	status := parts[0]
	prop := parts[1]

	parts = strings.SplitN(status, " ", 3)
	res.Version = parts[0]
	if res.StatusCode, err = strconv.Atoi(parts[1]); err != nil {
		return nil, err
	}
	res.Status = parts[2]

	pairs := strings.Split(prop, "\r\n")
	for _, pair := range pairs {
		parts = strings.SplitN(pair, ": ", 2)
		key := parts[0]
		value := parts[1]
		res.Header.Add(key, value)
	}

	res.Body = string(body)

	return res, nil
}
