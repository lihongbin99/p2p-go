package p2p_udp

import (
	"fmt"
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
		"udp_p2p_3389":  3389,
		"udp_p2p_speed": 13522,
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
	serverUDPAddr      string
	registerReloadUUID time.Time
}

func NewClientServer(writeChan chan msg.Message) *ClientServer {
	return &ClientServer{writeChan, RegisterError, "", time.Now()}
}

func (s *ClientServer) ChangeStatus(cId int32, sId int32, status byte) {
	if status == handle.Success {
		log.Trace(cId, sId, "send register request")
		s.writeChan <- &msg.UDPRegisterRequestMessage{Names: udpServerNames[:]}
	} else if status == handle.Error {
		log.Info(cId, sId, "change status register error")
		s.serverStatus = RegisterError
	}
}

func (s *ClientServer) Handle(cId int32, sId int32, message *io.Message) (handle bool) {
	switch message.Message.(type) {
	case *msg.UDPRegisterResponseMessage:
	case *msg.UDPNewConnectResponseMessage:
	default:
		return
	}

	handle = true
	switch m := message.Message.(type) {
	case *msg.UDPRegisterResponseMessage:
		if m.Status == msg.UDPRegisterSuccess {
			s.serverUDPAddr = config.ServerIp + ":" + strconv.Itoa(int(m.ServerUDPPort))
			log.Info(cId, sId, fmt.Sprintf("Register Success: server udp addr[%s]", s.serverUDPAddr))
			s.serverStatus = RegisterSuccess
		} else {
			s.serverStatus = Registering
			reloadUUID := time.Now()
			s.registerReloadUUID = reloadUUID
			// 等待一段时间后重试
			go func(cp *ClientServer, cReloadUUID time.Time) {
				time.Sleep(1 * time.Minute)
				if cp.serverStatus == Registering && cp.registerReloadUUID == cReloadUUID {
					cp.writeChan <- &msg.UDPRegisterRequestMessage{Names: udpServerNames[:]}
				}
			}(s, reloadUUID)
			log.Error(cId, sId, fmt.Errorf("register Error: %v", m.Msg))
		}
	case *msg.UDPNewConnectResponseMessage:
		go newConnect(s, m.Cid, m.Name, m.CIp+":"+strconv.Itoa(int(m.CPort)))
	default:
		handle = false
	}
	return
}

func newConnect(s *ClientServer, cId int32, name string, remoteAddrS string) {
	newId := getNewCid()
	appAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+strconv.Itoa(serverNameMap[name]))
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("resolveUDPAddr: %v", err))
		return
	}
	appConn, err := net.DialUDP("udp4", nil, appAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("dialUDP: %v", err))
		return
	}

	log.Info(cId, newId, remoteAddrS)
	// 发送探测包
	remoteAddr, err := net.ResolveUDPAddr("udp4", remoteAddrS)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("resolveUDPAddr: %v", err))
		return
	}
	tc, err := net.DialUDP("udp4", nil, remoteAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("dialUDP: %v", err))
		return
	}
	defer func() { _ = tc.Close() }()
	testUDP := io.NewUDPById(tc, newId)
	_, err = testUDP.WriteMessage(&msg.IgnoreMessage{})
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("writeMessage IgnoreMessage: %v", err))
		return
	}
	log.Trace(cId, newId, fmt.Sprintf("send IgnoreMessage to [%s]", remoteAddr))
	_ = testUDP.Close()

	localAddr, _ := net.ResolveUDPAddr("udp4", testUDP.LocalAddr().String())

	// 发送响应包
	serverAddr, err := net.ResolveUDPAddr("udp4", s.serverUDPAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("resolveUDPAddr: %v", err))
		return
	}
	ts, err := net.DialUDP("udp4", localAddr, serverAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("dialUDP: %v", err))
		return
	}
	defer func() { _ = ts.Close() }()
	testUDP = io.NewUDPById(ts, newId)
	_, err = testUDP.WriteMessage(&msg.UDPNewConnectResultRequestMessage{Cid: cId, Sid: newId})
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("writeMessage UDPNewConnectResultRequestMessage: %v", err))
		return
	}
	log.Trace(cId, newId, fmt.Sprintf("send message UDPNewConnectResultRequestMessage to [%s]", s.serverUDPAddr))
	_ = testUDP.Close()

	l, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("listenUDP: %v", err))
		return
	}
	udp := io.NewUDPById(l, newId)
	udp.Tid = cId
	log.Trace(udp.Tid, udp.Id, "ListenUDP success")

	_ = udp.SetReadDeadline(time.Now().Add(10 * time.Second))
	message, rd, err := udp.ReadMessageFromUDP()
	if err != nil {
		log.Error(udp.Tid, udp.Id, fmt.Errorf("read UDPNewConnectResultRequestMessage error: %v", err))
		return
	}
	_ = udp.SetReadDeadline(time.Time{})
	switch m := message.(type) {
	case *msg.UDPNewConnectResultRequestMessage:
		if m.Sid == udp.Id {
			if _, err := udp.WriteMessageToUDP(&msg.UDPNewConnectResultRequestMessage{Cid: udp.Tid}, rd); err != nil {
				log.Error(udp.Tid, udp.Id, fmt.Errorf("write UDPNewConnectResultRequestMessage error: %v", err))
				return
			}
		} else {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("sid error: %d", m.Sid))
			return
		}
	default:
		log.Error(udp.Tid, udp.Id, fmt.Errorf("message type no UDPNewConnectResultRequestMessage error: %v", message.ToByteBuf()))
		return
	}

	// 交换数据
	log.Info(udp.Tid, udp.Id, fmt.Sprintf("nat success: [%s]", remoteAddrS))
	go func(dest *net.UDPConn, src *net.UDPConn) {
		buf := make([]byte, 64*1024)
		for {
			readLength, err := src.Read(buf)
			if err != nil {
				log.Error(udp.Tid, udp.Id, fmt.Errorf("read from p2p error: %v", err))
				break
			}
			_, err = dest.Write(buf[:readLength])
			if err != nil {
				log.Error(udp.Tid, udp.Id, fmt.Errorf("write to app error: %v", err))
				break
			}
		}
	}(appConn, udp.UDPConn)

	go func(dest *net.UDPConn, src *net.UDPConn, rd *net.UDPAddr) {
		buf := make([]byte, 64*1024)
		for {
			readLength, err := src.Read(buf)
			if err != nil {
				log.Error(udp.Tid, udp.Id, fmt.Errorf("read from app error: %v", err))
				break
			}
			if readLength > 1464 {
				log.Error(udp.Id, udp.Tid, fmt.Errorf("long read: %d", readLength))
			}
			_, err = dest.WriteToUDP(buf[:readLength], rd)
			if err != nil {
				log.Error(udp.Tid, udp.Id, fmt.Errorf("write to p2p error: %v", err))
				break
			}
		}
	}(udp.UDPConn, appConn, rd)
}
