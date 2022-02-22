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
		//"udp_p2p_3389": 3389,
		//"udp_p2p_5938": 5938,
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
	case *msg.UDPTransferRequestMessage:
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
	case *msg.UDPTransferRequestMessage:
		go newTransfer(s, m.Sid, m.Name)
	default:
		handle = false
	}
	return
}

func newConnect(s *ClientServer, cId int32, name string, remoteAddrS string) {
	newId := getNewCid()

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
		_ = udp.Close()
		return
	}
	switch m := message.(type) {
	case *msg.UDPNewConnectResultRequestMessage:
		if m.Sid == udp.Id {
			if _, err := udp.WriteMessageToUDP(&msg.UDPNewConnectResultRequestMessage{Cid: udp.Tid}, rd); err != nil {
				log.Error(udp.Tid, udp.Id, fmt.Errorf("write UDPNewConnectResultRequestMessage error: %v", err))
				_ = udp.Close()
				return
			}
		} else {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("sid error: %d", m.Sid))
			_ = udp.Close()
			return
		}
	default:
		log.Error(udp.Tid, udp.Id, fmt.Errorf("message type no UDPNewConnectResultRequestMessage error: %v", message.ToByteBuf()))
		_ = udp.Close()
		return
	}
	_ = udp.SetReadDeadline(time.Time{})

	// udp 关流
	go goTimeOut(udp)

	appAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+strconv.Itoa(serverNameMap[name]))
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("resolveUDPAddr: %v", err))
		_ = udp.Close()
		return
	}
	appConn, err := net.DialUDP("udp4", nil, appAddr)
	if err != nil {
		log.Error(cId, newId, fmt.Errorf("dialUDP: %v", err))
		_ = udp.Close()
		return
	}

	// 交换数据
	log.Info(udp.Tid, udp.Id, fmt.Sprintf("nat success: [%s]", remoteAddrS))
	go a2a(appConn, udp.UDPConn, udp)
	go a2b(udp.UDPConn, appConn, rd, udp)
}

func newTransfer(s *ClientServer, sid int32, name string) {
	newId := getNewCid()
	log.Info(newId, sid, fmt.Sprintf("udp transfer [%s]", name))
	serverAddr, err := net.ResolveUDPAddr("udp4", s.serverUDPAddr)
	if err != nil {
		log.Error(newId, sid, fmt.Errorf("resolveUDPAddr: %v", err))
		return
	}
	u, err := net.DialUDP("udp4", nil, serverAddr)
	if err != nil {
		log.Error(newId, sid, fmt.Errorf("dialTCP: %v", err))
		return
	}
	udp := io.NewUDPById(u, newId)
	udp.Tid = sid

	appAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(serverNameMap[name]))
	if err != nil {
		log.Error(udp.Id, udp.Tid, fmt.Errorf("resolveTCPAddr: %v", err))
		_, _ = udp.WriteMessage(&msg.UDPTransferResponseMessage{Message: err.Error()})
		_ = udp.Close()
		return
	}
	appConn, err := net.DialUDP("udp4", nil, appAddr)
	if err != nil {
		log.Error(udp.Id, udp.Tid, fmt.Errorf("dialTCP: %v", err))
		_, _ = udp.WriteMessage(&msg.UDPTransferResponseMessage{Message: err.Error()})
		_ = udp.Close()
		return
	}

	_, err = udp.WriteMessage(&msg.UDPTransferResponseMessage{Sid: sid})
	if err != nil {
		log.Error(udp.Id, udp.Tid, fmt.Errorf("writeMessage UDPTransferResponseMessage: %v", err))
		_ = udp.Close()
		_ = appConn.Close()
		return
	}

	// udp 关流
	go goTimeOut(udp)

	// 开始传输数据
	go a2a2(udp.UDPConn, appConn, udp)
	go a2a(appConn, udp.UDPConn, udp)
}

func a2a2(dest *net.UDPConn, src *net.UDPConn, udp *io.UDP) {
	defer func() {
		_ = src.Close()
		_ = dest.Close()
	}()
	buf := make([]byte, 64*1024)
	for {
		readLength, err := src.Read(buf)
		if udp.TimeOut {
			break
		}
		if err != nil {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("read from p2p error: %v", err))
			break
		}
		udp.LastTransferTime = time.Now()
		if _, err = dest.Write(buf[:readLength]); err != nil {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("write to app error: %v", err))
			break
		}
	}
}

func a2a(dest *net.UDPConn, src *net.UDPConn, udp *io.UDP) {
	defer func() {
		_ = src.Close()
		_ = dest.Close()
	}()
	buf := make([]byte, 64*1024)
	for {
		readLength, err := src.Read(buf)
		if udp.TimeOut {
			break
		}
		if err != nil {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("read from p2p error: %v", err))
			break
		}
		udp.LastTransferTime = time.Now()
		if _, err = dest.Write(buf[:readLength]); err != nil {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("write to app error: %v", err))
			break
		}
	}
}

func a2b(dest *net.UDPConn, src *net.UDPConn, rd *net.UDPAddr, udp *io.UDP) {
	defer func() {
		_ = src.Close()
		_ = dest.Close()
	}()
	buf := make([]byte, 64*1024)
	for {
		readLength, err := src.Read(buf)
		if udp.TimeOut {
			break
		}
		if err != nil {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("read from app error: %v", err))
			break
		}
		udp.LastTransferTime = time.Now()
		if readLength > 1464 {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("long read: %d", readLength))
		}
		if _, err = dest.WriteToUDP(buf[:readLength], rd); err != nil {
			log.Error(udp.Tid, udp.Id, fmt.Errorf("write to p2p error: %v", err))
			break
		}
	}
}
