package p2p_tcp

import (
	"fmt"
	sio "io"
	"net"
	"p2p-go/client/config"
	"p2p-go/common/handle"
	"p2p-go/common/io"
	"p2p-go/common/msg"
	"strconv"
	"sync"
	"time"
)

const (
	_ byte = iota
	Accessing
	AccessError
	AccessSuccess
)

var (
	clientNameMap = map[string]int{
		"tcp_p2p_3389":  4567,
		"tcp_p2p_speed": 13521,
	}
	tcpClientNames = make([]string, 0)
)

func init() {
	for name := range clientNameMap {
		tcpClientNames = append(tcpClientNames, name)
	}
}

type ClientClient struct {
	writeChan chan msg.Message

	clientStatus     map[string]byte
	accessReloadLock sync.Mutex
	accessReloadUUID time.Time

	serverTCPAddr string
}

func NewClientClient(writeChan chan msg.Message) (result *ClientClient) {
	clientStatus := make(map[string]byte)
	for _, name := range tcpClientNames {
		clientStatus[name] = AccessError
	}
	result = &ClientClient{
		writeChan,
		clientStatus,
		sync.Mutex{},
		time.Now(),
		config.ServerIp + ":" + strconv.Itoa(config.ServerPort),
	}
	for name, port := range clientNameMap {
		thisAddr, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:"+strconv.Itoa(port))
		if err != nil {
			log.Fatal(fmt.Errorf("resolveTCPAddr: %v", err))
		}

		listener, err := net.ListenTCP("tcp4", thisAddr)
		if err != nil {
			log.Fatal(fmt.Errorf("listenTCP: %v", err))
		}
		log.Info(0, 0, fmt.Sprintf("start server success: %s", thisAddr.String()))

		go func(l *net.TCPListener, c *ClientClient, name string) {
			for {
				conn, err := l.AcceptTCP()
				if err != nil {
					log.Error(0, 0, fmt.Errorf("acceptTCP: %v", err))
					continue
				}
				go createConnect(c, conn, name)
			}
		}(listener, result, name)
	}
	return
}

func createConnect(c *ClientClient, appConn *net.TCPConn, name string) {
	defer func() { _ = appConn.Close() }()
	newId := getNewCid()
	log.Trace(newId, 0, "new connect")
	serverAddr, err := net.ResolveTCPAddr("tcp4", c.serverTCPAddr)
	if err != nil {
		log.Error(newId, 0, fmt.Errorf("resolveTCPAddr: %v", err))
		return
	}
	t, err := net.DialTCP("tcp4", nil, serverAddr)
	if err != nil {
		log.Error(newId, 0, fmt.Errorf("dialTCP: %v", err))
		return
	}
	defer func() { _ = t.Close() }()
	tcp := io.NewTCPById(t, newId)

	// 创建 p2p 连接
	_, err = tcp.WriteMessage(&msg.TCPNewConnectRequestMessage{Cid: tcp.Id, Name: name})
	if err != nil {
		log.Error(tcp.Id, 0, fmt.Errorf("writeMessage TCPNewConnectRequestMessage: %v", err))
		return
	}
	log.Trace(tcp.Id, 0, fmt.Sprintf("send message TCPNewConnectRequestMessage [%s]", c.serverTCPAddr))

	_ = tcp.SetReadDeadline(time.Now().Add(10 * time.Second))
	message := tcp.ReadMessage()
	if message.Err != nil {
		log.Error(tcp.Id, 0, fmt.Errorf("read TCPNewConnectResultResponseMessage error: %v", message.Err))
		return
	}
	_ = tcp.SetReadDeadline(time.Time{})
	_ = tcp.Close()

	var m *msg.TCPNewConnectResultResponseMessage = nil
	switch t := message.Message.(type) {
	case *msg.TCPNewConnectResultResponseMessage:
		m = t
	default:
		log.Error(tcp.Id, 0, fmt.Errorf("read message type no TCPNewConnectResultResponseMessage: %v", message.Message.ToByteBuf()))
		return
	}

	remoteAddrS := m.SIp + ":" + strconv.Itoa(int(m.SPort))
	log.Info(tcp.Id, m.Sid, fmt.Sprintf("remote addr : [%s]", remoteAddrS))
	remoteAddr, err := net.ResolveTCPAddr("tcp4", remoteAddrS)
	if err != nil {
		log.Error(tcp.Id, m.Sid, fmt.Errorf("resolveTCPAddr: %v", err))
		return
	}
	rt, err := net.DialTCP("tcp4", nil, remoteAddr)
	if err != nil {
		log.Error(tcp.Id, m.Sid, fmt.Errorf("dialTCP: %v", err))
		return
	}
	defer func() { _ = rt.Close() }()
	tcp = io.NewTCPById(rt, tcp.Id)
	tcp.Tid = m.Sid

	endDelayTime := time.Time{}
	startDelayTime := time.Now()
	_, err = tcp.WriteMessage(&msg.TCPNewConnectResultRequestMessage{Sid: tcp.Tid})
	if err != nil {
		log.Error(tcp.Id, tcp.Tid, fmt.Errorf("write TCPNewConnectResultRequestMessage error: %v", err))
		_ = tcp.Close()
		return
	}
	log.Trace(tcp.Id, tcp.Tid, "write to client-server TCPNewConnectResultRequestMessage")
	_ = tcp.SetReadDeadline(time.Now().Add(10 * time.Second))
	if message = tcp.ReadMessage(); message.Err != nil {
		log.Error(tcp.Id, tcp.Tid, fmt.Errorf("read TCPNewConnectResultRequestMessage error: %v", message.Err))
		_ = tcp.Close()
		return
	} else {
		switch m := message.Message.(type) {
		case *msg.TCPNewConnectResultRequestMessage:
			if m.Cid != tcp.Id {
				log.Error(tcp.Tid, tcp.Id, fmt.Errorf("cid error: %d", m.Cid))
				_ = tcp.Close()
				return
			}
			endDelayTime = time.Now()
		default:
			log.Error(tcp.Id, tcp.Tid, fmt.Errorf("read message type no TCPNewConnectResultRequestMessage error: %v", message.Message.ToByteBuf()))
			_ = tcp.Close()
			return
		}
	}
	_ = tcp.SetReadDeadline(time.Time{})

	// 开始交换数据
	log.Info(tcp.Id, tcp.Tid, fmt.Sprintf("nat success: [%s], delay: %dms", remoteAddrS, endDelayTime.Sub(startDelayTime)/time.Millisecond))
	go func() {
		_, _ = sio.Copy(appConn, tcp)
	}()
	_, _ = sio.Copy(tcp, appConn)
}

func (c *ClientClient) ChangeStatus(cId int32, sId int32, status byte) {
	if status == handle.Success {
		c.flushNames(true)
		log.Trace(cId, sId, "send access request")
	} else if status == handle.Error {
		for k := range c.clientStatus {
			c.clientStatus[k] = AccessError
		}
		log.Info(cId, sId, "change all status access error")
	}
}

func (c *ClientClient) Handle(cId int32, sId int32, message *io.Message) (handle bool) {
	switch message.Message.(type) {
	case *msg.TCPAccessResponseMessage:
	case *msg.TCPAccessCloseResponseMessage:
	case *msg.TCPAccessFlushResponseMessage:
	default:
		return
	}

	handle = true
	switch m := message.Message.(type) {
	case *msg.TCPAccessResponseMessage:
		for _, name := range m.SuccessNames {
			c.clientStatus[name] = AccessSuccess
		}
		c.flushNames(false)
		if len(m.SuccessNames) > 0 {
			log.Info(cId, sId, fmt.Sprintf("access success: [%v]", m.SuccessNames))
		}
	case *msg.TCPAccessCloseResponseMessage:
		for _, name := range m.CloseNames {
			c.clientStatus[name] = AccessError
		}
		log.Error(cId, sId, fmt.Errorf("access close%v", tcpClientNames))
		c.flushNames(false)
	case *msg.TCPAccessFlushResponseMessage:
		log.Trace(0, 0, "flush name")
		c.flushNames(true)
	default:
		handle = false
	}
	return
}

func (c *ClientClient) flushNames(now bool) {
	c.accessReloadLock.Lock()
	defer c.accessReloadLock.Unlock()
	names := make([]string, 0)
	for name, status := range c.clientStatus {
		if status != AccessSuccess {
			c.clientStatus[name] = Accessing
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return
	}

	reloadUUID := time.Now()
	c.accessReloadUUID = reloadUUID
	go func(cp *ClientClient, cReloadUUID time.Time, names []string) {
		if !now {
			time.Sleep(1 * time.Minute)
		}
		c.accessReloadLock.Lock()
		defer c.accessReloadLock.Unlock()
		flushNames := make([]string, 0)
		if cp.accessReloadUUID == cReloadUUID {
			for _, name := range names {
				if cp.clientStatus[name] == Accessing {
					flushNames = append(flushNames, name)
				}
			}
		}

		if len(flushNames) > 0 {
			cp.writeChan <- &msg.TCPAccessRequestMessage{Names: flushNames[:]}
		}
	}(c, reloadUUID, names)
}