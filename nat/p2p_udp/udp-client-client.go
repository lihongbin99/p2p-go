package p2p_udp

import (
	"fmt"
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
		"udp_p2p_3389": 4567,
		"udp_p2p_5938": 5937,
		//"udp_p2p_speed": 13521,
	}
	udpClientNames = make([]string, 0)
)

func init() {
	for name := range clientNameMap {
		udpClientNames = append(udpClientNames, name)
	}
}

type ClientClient struct {
	writeChan chan msg.Message

	clientStatus     map[string]byte
	accessReloadLock sync.Mutex
	accessReloadUUID time.Time

	serverUDPAddr string

	addrCache map[string]*LinkCache
	addrTable map[string]*io.UDP

	localUDPListener map[string]*net.UDPConn
	localUDPLock     sync.Mutex

	lock sync.Mutex
}

type LinkCache struct {
	*net.UDPAddr
	bufList [][]byte
}

func NewClientClient(writeChan chan msg.Message) (result *ClientClient) {
	clientStatus := make(map[string]byte)
	for _, name := range udpClientNames {
		clientStatus[name] = AccessError
	}
	result = &ClientClient{
		writeChan,
		clientStatus,
		sync.Mutex{},
		time.Now(),
		"",
		make(map[string]*LinkCache, 8),
		make(map[string]*io.UDP, 8),
		make(map[string]*net.UDPConn, 8),
		sync.Mutex{},
		sync.Mutex{},
	}
	for name, port := range clientNameMap {
		thisAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(port))
		if err != nil {
			log.Fatal(fmt.Errorf("resolveUDPAddr: %v", err))
		}

		listener, err := net.ListenUDP("udp4", thisAddr)
		if err != nil {
			log.Fatal(fmt.Errorf("listenUDP: %v", err))
		}
		result.localUDPListener[name] = listener
		log.Info(0, 0, fmt.Sprintf("start server success: %s", thisAddr.String()))

		go func(l *net.UDPConn, c *ClientClient, name string) {
			buf := make([]byte, 64*1024)
			for {
				readLength, localAddr, err := l.ReadFromUDP(buf)
				if err != nil {
					log.Error(0, 0, fmt.Errorf("readFromUDP: %v", err))
					continue
				}
				clientRunNewMessage(c, buf, readLength, localAddr, name)
			}
		}(listener, result, name)
	}
	return
}

func clientRunNewMessage(c *ClientClient, buf []byte, readLength int, localAddr *net.UDPAddr, name string) {
	// 连接创建完成, 直接交换数据
	if clientExchangeData(c, buf, readLength, localAddr) {
		return
	}

	c.lock.Lock()
	c.lock.Unlock()
	if clientExchangeData(c, buf, readLength, localAddr) {
		return
	}

	if c.clientStatus[name] != AccessSuccess {
		log.Error(0, 0, fmt.Errorf("client status error"))
		return
	}

	// 连接创建中, 增加缓存数据
	bufCache := make([]byte, readLength)
	copy(bufCache, buf[:readLength])
	if linkCache, ok := c.addrCache[localAddr.String()]; ok {
		linkCache.bufList = append(linkCache.bufList, bufCache)
	} else {
		linkCache = &LinkCache{localAddr, make([][]byte, 1)}
		linkCache.bufList[0] = bufCache
		c.addrCache[localAddr.String()] = linkCache
		go func() {
			if !createConnect(c, localAddr, name) {
				delete(c.addrCache, localAddr.String())
			}
		}()
	}
}

func clientExchangeData(c *ClientClient, buf []byte, readLength int, localAddr *net.UDPAddr) bool {
	if remoteUDP, ok := c.addrTable[localAddr.String()]; ok {
		if _, err := remoteUDP.Write(buf[:readLength]); err != nil {
			log.Error(remoteUDP.Id, remoteUDP.Tid, fmt.Errorf("exchange: %v", err))
			delete(c.addrTable, localAddr.String())
		}
		return true
	}
	return false
}

func createConnect(c *ClientClient, localAddr *net.UDPAddr, name string) (success bool) {
	startCreateTime := time.Now()
	newId := getNewCid()
	log.Trace(newId, 0, "new connect")
	serverAddr, err := net.ResolveUDPAddr("udp4", c.serverUDPAddr)
	if err != nil {
		log.Error(newId, 0, fmt.Errorf("resolveUDPAddr: %v", err))
		return
	}
	u, err := net.DialUDP("udp4", nil, serverAddr)
	if err != nil {
		log.Error(newId, 0, fmt.Errorf("dialUDP: %v", err))
		return
	}
	defer func() { _ = u.Close() }()
	udp := io.NewUDPById(u, newId)
	udpAddr, _ := net.ResolveUDPAddr("udp4", udp.LocalAddr().String())

	// 创建 p2p 连接
	_, err = udp.WriteMessage(&msg.UDPNewConnectRequestMessage{Cid: udp.Id, Name: name})
	if err != nil {
		log.Error(udp.Id, 0, fmt.Errorf("writeMessage UDPNewConnectRequestMessage: %v", err))
		return
	}
	log.Trace(udp.Id, 0, fmt.Sprintf("send message UDPNewConnectRequestMessage [%s]", c.serverUDPAddr))

	_ = udp.SetReadDeadline(time.Now().Add(10 * time.Second))
	message, _, err := udp.ReadMessageFromUDP()
	if err != nil {
		log.Error(udp.Id, 0, fmt.Errorf("read UDPNewConnectResultResponseMessage error: %v", err))
		return
	}
	_ = udp.SetReadDeadline(time.Time{})

	switch m := message.(type) {
	case *msg.IgnoreMessage:
		m.ToByteBuf()
		log.Trace(udp.Id, 0, "read IgnoreMessage, start reread UDPNewConnectResultResponseMessage")
		_ = udp.SetReadDeadline(time.Now().Add(10 * time.Second))
		message, _, err = udp.ReadMessageFromUDP()
		if err != nil {
			log.Error(udp.Id, 0, fmt.Errorf("reread UDPNewConnectResultResponseMessage error: %v", err))
			return
		}
		_ = udp.SetReadDeadline(time.Time{})
	}
	_ = udp.Close()

	var m *msg.UDPNewConnectResultResponseMessage = nil
	switch t := message.(type) {
	case *msg.UDPNewConnectResultResponseMessage:
		m = t
	default:
		log.Error(udp.Id, 0, fmt.Errorf("read message type no UDPNewConnectResultResponseMessage: %v", message.ToByteBuf()))
		return
	}

	remoteAddrS := m.SIp + ":" + strconv.Itoa(int(m.SPort))
	log.Info(udp.Id, m.Sid, fmt.Sprintf("remote addr : [%s]", remoteAddrS))
	remoteAddr, err := net.ResolveUDPAddr("udp4", remoteAddrS)
	if err != nil {
		log.Error(udp.Id, m.Sid, fmt.Errorf("resolveUDPAddr: %v", err))
		return
	}
	ru, err := net.DialUDP("udp4", udpAddr, remoteAddr)
	if err != nil {
		log.Error(udp.Id, m.Sid, fmt.Errorf("dialUDP: %v", err))
		return
	}
	udp = io.NewUDPById(ru, udp.Id)
	udp.Tid = m.Sid

	endDelayTime := time.Time{}
	startDelayTime := time.Now()
	_, err = udp.WriteMessage(&msg.UDPNewConnectResultRequestMessage{Sid: udp.Tid})
	if err != nil {
		log.Error(udp.Id, udp.Tid, fmt.Errorf("write UDPNewConnectResultRequestMessage error: %v", err))
		_ = udp.Close()
		return
	}
	log.Trace(udp.Id, udp.Tid, "write to client-server UDPNewConnectResultRequestMessage")
	_ = udp.SetReadDeadline(time.Now().Add(10 * time.Second))
	if message, _, err = udp.ReadMessageFromUDP(); err != nil {
		log.Error(udp.Id, udp.Tid, fmt.Errorf("read UDPNewConnectResultRequestMessage error: %v", err))
		_ = udp.Close()
		return
	} else {
		switch m := message.(type) {
		case *msg.UDPNewConnectResultRequestMessage:
			if m.Cid != udp.Id {
				log.Error(udp.Tid, udp.Id, fmt.Errorf("cid error: %d", m.Cid))
				_ = udp.Close()
				return
			}
			endDelayTime = time.Now()
		default:
			log.Error(udp.Id, udp.Tid, fmt.Errorf("read message type no UDPNewConnectResultRequestMessage error: %v", message.ToByteBuf()))
			_ = udp.Close()
			return
		}
	}
	_ = udp.SetReadDeadline(time.Time{})

	endCreateTime := time.Now()
	// 开始交换数据
	log.Info(udp.Id, udp.Tid, fmt.Sprintf("nat success: [%s], penetrate: %dms, delay: %dms", remoteAddrS, endCreateTime.Sub(startCreateTime)/time.Millisecond, endDelayTime.Sub(startDelayTime)/time.Millisecond))

	c.lock.Lock()
	c.lock.Unlock()
	linkCache := c.addrCache[localAddr.String()]
	delete(c.addrCache, localAddr.String())
	for _, buf := range linkCache.bufList {
		if _, err := udp.Write(buf); err != nil {
			log.Error(udp.Id, udp.Tid, fmt.Errorf("fast send message error: %v", err))
			_ = udp.Close()
			return
		}
	}

	c.addrTable[localAddr.String()] = udp
	// 返回数据给软件
	go func(c *ClientClient, udp *io.UDP, localAddr *net.UDPAddr, name string) {
		buf := make([]byte, 64*1024)
		for {
			readLength, err := udp.Read(buf)
			if err != nil {
				log.Error(udp.Id, udp.Tid, fmt.Errorf("read from p2p error: %v", err))
				break
			}
			if readLength > 1464 {
				log.Error(udp.Id, udp.Tid, fmt.Errorf("long read: %d", readLength))
			}
			_, err = c.writeToLocal(buf[:readLength], name, localAddr)
			if err != nil {
				log.Error(udp.Id, udp.Tid, fmt.Errorf("write to local error: %v", err))
				break
			}
		}
	}(c, udp, localAddr, name)
	return true
}

func (c *ClientClient) writeToLocal(buf []byte, name string, addr *net.UDPAddr) (writeLength int, err error) {
	c.localUDPLock.Lock()
	defer c.localUDPLock.Unlock()
	writeLength, err = c.localUDPListener[name].WriteToUDP(buf, addr)
	return
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
	case *msg.UDPAccessResponseMessage:
	case *msg.UDPAccessCloseResponseMessage:
	case *msg.UDPAccessFlushResponseMessage:
	default:
		return
	}

	handle = true
	switch m := message.Message.(type) {
	case *msg.UDPAccessResponseMessage:
		c.serverUDPAddr = config.ServerIp + ":" + strconv.Itoa(int(m.ServerUDPPort))
		for _, name := range m.SuccessNames {
			c.clientStatus[name] = AccessSuccess
		}
		c.flushNames(false)
		if len(m.SuccessNames) > 0 {
			log.Info(cId, sId, fmt.Sprintf("server udp addr: %s, access: [%v]", c.serverUDPAddr, m.SuccessNames))
		}
	case *msg.UDPAccessCloseResponseMessage:
		for _, name := range m.CloseNames {
			c.clientStatus[name] = AccessError
		}
		log.Error(cId, sId, fmt.Errorf("access close%v", udpClientNames))
		c.flushNames(false)
	case *msg.UDPAccessFlushResponseMessage:
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
			cp.writeChan <- &msg.UDPAccessRequestMessage{Names: flushNames[:]}
		}
	}(c, reloadUUID, names)
}
