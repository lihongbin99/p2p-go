package p2p_tcp

import (
	"fmt"
	sio "io"
	"math/rand"
	"net"
	"p2p-go/client/config"
	"p2p-go/common/handle"
	"p2p-go/common/io"
	"p2p-go/common/msg"
	"strconv"
	"time"
)

const (
	_ byte = iota
	Registering
	RegisterError
	RegisterSuccess
)

var (
	serverNameMap = map[string]int{
		"tcp_p2p_3389": 3389,
		"tcp_p2p_5938": 5938,
		//"tcp_p2p_speed": 13522,
	}
	udpServerNames = make([]string, 0)
)

func init() {
	for name := range serverNameMap {
		udpServerNames = append(udpServerNames, name)
	}
}

type ClientServer struct {
	writeChan          chan msg.Message
	serverStatus       byte
	registerReloadUUID time.Time

	serverTCPAddr string
}

func NewClientServer(writeChan chan msg.Message) *ClientServer {
	return &ClientServer{
		writeChan,
		RegisterError,
		time.Now(),
		config.ServerIp + ":" + strconv.Itoa(config.ServerPort),
	}
}

func (s *ClientServer) ChangeStatus(cId int32, sId int32, status byte) {
	if status == handle.Success {
		log.Trace(cId, sId, "send register request")
		s.writeChan <- &msg.TCPRegisterRequestMessage{Names: udpServerNames[:]}
	} else if status == handle.Error {
		log.Info(cId, sId, "change status register error")
		s.serverStatus = RegisterError
	}
}

func (s *ClientServer) Handle(cId int32, sId int32, message *io.Message) (handle bool) {
	switch message.Message.(type) {
	case *msg.TCPRegisterResponseMessage:
	case *msg.TCPNewConnectResponseMessage:
	default:
		return
	}

	handle = true
	switch m := message.Message.(type) {
	case *msg.TCPRegisterResponseMessage:
		if m.Status == msg.TCPRegisterSuccess {
			log.Info(cId, sId, fmt.Sprintf("Register Success"))
			s.serverStatus = RegisterSuccess
		} else {
			s.serverStatus = Registering
			reloadUUID := time.Now()
			s.registerReloadUUID = reloadUUID
			// 等待一段时间后重试
			go func(cp *ClientServer, cReloadUUID time.Time) {
				time.Sleep(1 * time.Minute)
				if cp.serverStatus == Registering && cp.registerReloadUUID == cReloadUUID {
					cp.writeChan <- &msg.TCPRegisterRequestMessage{Names: udpServerNames[:]}
				}
			}(s, reloadUUID)
			log.Error(cId, sId, fmt.Errorf("register Error: %v", m.Msg))
		}
	case *msg.TCPNewConnectResponseMessage:
		go newConnect(s, m.Cid, m.Name, m.CIp+":"+strconv.Itoa(int(m.CPort)))
	default:
		handle = false
	}
	return
}

func newConnect(s *ClientServer, cId int32, name string, remoteAddrS string) {
	newId := getNewCid()
	appAddr, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(serverNameMap[name]))
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("resolveTCPAddr: %v", err))
		return
	}
	appConn, err := net.DialTCP("tcp4", nil, appAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("dialTCP: %v", err))
		return
	}

	log.Info(cId, newId, remoteAddrS)
	localAddr, _ := net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(rand.Intn(10000)+50000))

	dialer := net.Dialer{Timeout: 10 * time.Millisecond, LocalAddr: localAddr}
	// 发送探测包
	tc, err := dialer.Dial("tcp4", remoteAddrS)
	if err == nil {
		_ = tc.Close()
	}

	// 发送响应包
	serverAddr, err := net.ResolveTCPAddr("tcp4", s.serverTCPAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("resolveTCPAddr: %v", err))
		return
	}
	ts, err := net.DialTCP("tcp4", localAddr, serverAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("dialTCP: %v", err))
		return
	}
	defer func() { _ = ts.Close() }()
	testTCP := io.NewTCPById(ts, newId)
	_, err = testTCP.WriteMessage(&msg.TCPNewConnectResultRequestMessage{Cid: cId, Sid: newId})
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("writeMessage TCPNewConnectResultRequestMessage: %v", err))
		return
	}
	log.Trace(cId, newId, fmt.Sprintf("send message TCPNewConnectResultRequestMessage to [%s]", s.serverTCPAddr))
	_ = testTCP.Close()

	l, err := net.ListenTCP("tcp4", localAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("listenTCP: %v", err))
		return
	}
	defer func() { _ = l.Close() }()
	_ = l.SetDeadline(time.Now().Add(10 * time.Second))
	tp, err := l.AcceptTCP()
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("AcceptTCP: %v", err))
		return
	}
	defer func() { _ = tp.Close() }()
	_ = l.SetDeadline(time.Time{})

	tcp := io.NewTCPById(tp, newId)
	tcp.Tid = cId
	log.Trace(tcp.Tid, tcp.Id, "ListenTCP success")

	_ = tcp.SetReadDeadline(time.Now().Add(10 * time.Second))
	message := tcp.ReadMessage()
	if message.Err != nil {
		log.Error(tcp.Tid, tcp.Id, fmt.Errorf("read TCPNewConnectResultRequestMessage error: %v", message.Err))
		return
	}
	_ = tcp.SetReadDeadline(time.Time{})
	switch m := message.Message.(type) {
	case *msg.TCPNewConnectResultRequestMessage:
		if m.Sid == tcp.Id {
			if _, err := tcp.WriteMessage(&msg.TCPNewConnectResultRequestMessage{Cid: tcp.Tid}); err != nil {
				log.Error(tcp.Tid, tcp.Id, fmt.Errorf("write TCPNewConnectResultRequestMessage error: %v", err))
				return
			}
		} else {
			log.Error(tcp.Tid, tcp.Id, fmt.Errorf("sid error: %d", m.Sid))
			return
		}
	default:
		log.Error(tcp.Tid, tcp.Id, fmt.Errorf("message type no TCPNewConnectResultRequestMessage error: %v", message.Message.ToByteBuf()))
		return
	}

	// 交换数据
	log.Info(tcp.Tid, tcp.Id, fmt.Sprintf("nat success: [%s]", remoteAddrS))
	go func() {
		_, _ = sio.Copy(appConn, tcp)
	}()
	_, _ = sio.Copy(tcp, appConn)
}
