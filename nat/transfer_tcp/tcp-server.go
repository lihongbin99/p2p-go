package transfer_tcp

import (
	"fmt"
	"net"
	"p2p-go/common/id"
	"p2p-go/common/io"
	"p2p-go/common/msg"
	"sync"
)

type WriteChanMessage struct {
	cId       int32
	sId       int32
	writeChan chan msg.Message
}

type Server struct {
	writeChanMap map[int32]*WriteChanMessage
	lock         sync.Mutex

	listenMap map[int32][]*ListenCache // Map<sId, Map<Id, Array<ListenCache> >
	connMap   map[int32]*net.TCPConn
}

type ListenCache struct {
	*net.TCPListener
	id      int32
	sId     int32
	isClose bool
}

func NewServer() *Server {
	return &Server{
		make(map[int32]*WriteChanMessage),
		sync.Mutex{},
		make(map[int32][]*ListenCache),
		make(map[int32]*net.TCPConn),
	}
}

func (s *Server) NewTCP(tcp *io.TCP, message msg.Message) (re bool) {
	switch message.(type) {
	case *msg.TCPTunNewConnectResponseMessage:
	default:
		return
	}
	defer func() { _ = tcp.Close() }()

	switch m := message.(type) {
	case *msg.TCPTunNewConnectResponseMessage:
		if conn, ok := s.connMap[m.Sid]; ok {
			if m.Message != "" {
				log.Error(m.Cid, m.Sid, fmt.Errorf("tcp tun new connect error, message: %s", m.Message))
				_ = conn.Close()
			} else {
				log.Info(m.Cid, m.Sid, "tcp tun success")
				// 传输数据
				targetTCP := io.NewTCP(conn)
				go goTimeOut(tcp)
				go goTimeOut(targetTCP)
				go transferData(tcp, targetTCP)
				transferData(targetTCP, tcp)
			}
		} else {
			log.Error(m.Cid, m.Sid, fmt.Errorf("tcp tun no find sid"))
		}
	}
	return true
}

func (s *Server) NewConnect(_ int32, sId int32, writeChan chan msg.Message) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.writeChanMap[sId] = &WriteChanMessage{0, sId, writeChan}
}

func (s *Server) ConnectSuccess(cId int32, sId int32, _ chan msg.Message) {
	if writeChan, ok := s.writeChanMap[sId]; ok {
		writeChan.cId = cId
	}
}

func (s *Server) CloseConnect(_ int32, sId int32) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if listenCache, ok := s.listenMap[sId]; ok {
		for _, lc := range listenCache {
			lc.isClose = true
			_ = lc.Close()
		}
	}

	delete(s.writeChanMap, sId)
}

func (s *Server) Handle(cId int32, sId int32, ip string, port uint16, message *io.Message) (handle bool, re bool) {
	switch message.Message.(type) {
	case *msg.TCPTunRegisterRequestMessage:
	case *msg.TCPTunNewConnectResponseMessage:
	default:
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	writeChanMessage, ok := s.writeChanMap[sId]
	if !ok {
		log.Error(cId, sId, fmt.Errorf("no find write chan [%s:%d]", ip, port))
		return
	}

	handle = true
	switch m := message.Message.(type) {
	case *msg.TCPTunRegisterRequestMessage:
		tuns := BufToTun(m.Buf)
		registerSuccess := make([]int32, 0)
		for tunId, tun := range tuns {
			listenAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", tun.remoteAddr, tun.remotePort))
			if err != nil {
				log.Error(cId, sId, fmt.Errorf("resolve TCP addr error: %v", err))
				continue
			}
			listenTCP, err := net.ListenTCP("tcp4", listenAddr)
			if err != nil {
				log.Error(cId, sId, fmt.Errorf("listen TCP error: %v", err))
				continue
			}
			var listenArray []*ListenCache
			if listenArray, ok = s.listenMap[sId]; !ok {
				listenArray = make([]*ListenCache, 0)
			}
			listenCache := &ListenCache{TCPListener: listenTCP, id: tunId, sId: sId}
			listenArray = append(listenArray, &ListenCache{TCPListener: listenTCP})
			s.listenMap[sId] = listenArray
			go s.newListen(listenCache)
			log.Info(cId, sId, fmt.Sprintf("tcp tun [%s:%d] -> [%s:%d]", tun.remoteAddr, tun.remotePort, tun.localAddr, tun.localPort))

			registerSuccess = append(registerSuccess, tunId)
		}

		if len(registerSuccess) > 0 {
			writeChanMessage.writeChan <- &msg.TCPTunRegisterResponseMessage{Ids: registerSuccess}
		}
	case *msg.TCPTunNewConnectResponseMessage:
		if conn, ok := s.connMap[m.Sid]; ok {
			log.Error(cId, sId, fmt.Errorf("tcp tun new connect error, Cid[%d], Sid[%d], message: %s", m.Cid, m.Sid, m.Message))
			_ = conn.Close()
		}
	default:
		handle = false
	}
	return
}

func (s *Server) newListen(listen *ListenCache) {
	defer func() { _ = listen.Close() }()
	for {
		tcp, err := listen.AcceptTCP()
		if err != nil {
			if !listen.isClose {
				log.Error(0, 0, err)
			}
			break
		}
		newId := id.NewId()
		log.Trace(0, newId, fmt.Sprintf("tcp tun new accept form: %s -> %s", tcp.RemoteAddr().String(), listen.Addr().String()))
		s.connMap[newId] = tcp
		if writeChanMessage, ok := s.writeChanMap[listen.sId]; ok {
			writeChanMessage.writeChan <- &msg.TCPTunNewConnectRequestMessage{Id: listen.id, Sid: newId}
		}
	}
}
