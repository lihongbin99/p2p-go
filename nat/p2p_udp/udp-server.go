package p2p_udp

import (
	"fmt"
	"net"
	"p2p-go/common/io"
	"p2p-go/common/msg"
	"strconv"
	"strings"
	"sync"
)

type WriteChanMessage struct {
	cId       int32
	sId       int32
	writeChan chan msg.Message
}

var (
	udpPort uint16 = 13520
)

type Server struct {
	writeChanMap      map[int32]*WriteChanMessage
	registerIdMap     map[int32][]string         // Map<sId, List<sName> >
	registerNameMap   map[string]int32           // Map<sName, sId>
	accessIdMap       map[int32]map[string]int32 // Map<cId, Map<cName, sId> >
	accessNameMap     map[string]map[int32]int32 // Map<cName, Map<cId, sId> >
	hasErrorAccessMap map[int32]interface{}      // Map<cId, ant>

	p2pMap map[int32]*net.UDPAddr // Map<cId, cRemoteAddr>
	lock   sync.Mutex
}

func NewServer() (result *Server) {
	result = &Server{
		make(map[int32]*WriteChanMessage),
		make(map[int32][]string),
		make(map[string]int32),
		make(map[int32]map[string]int32),
		make(map[string]map[int32]int32),
		make(map[int32]interface{}),
		make(map[int32]*net.UDPAddr),
		sync.Mutex{},
	}
	// 开启 UDP 服务
	thisAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(int(udpPort)))
	if err != nil {
		log.Fatal(fmt.Errorf("resolveUDPAddr: %v", err))
	}

	listener, err := net.ListenUDP("udp4", thisAddr)
	if err != nil {
		log.Fatal(fmt.Errorf("listenUDP: %v", err))
	}
	l := io.NewUDP(listener)
	log.Info(0, 0, fmt.Sprintf("start server success: [%s]", thisAddr.String()))

	go func(udp *io.UDP, s *Server) {
		for {
			if message, remoteAddr, err := udp.ReadMessageFromUDP(); err != nil {
				log.Warn(0, 0, fmt.Sprintf("ReadMessageFromUDP: %v", err))
			} else {
				if wm, wra, w := s.serverRunNewMessage(message, remoteAddr); w {
					if _, err = udp.WriteMessageToUDP(wm, wra); err != nil {
						log.Warn(0, 0, fmt.Sprintf("WriteMessageToUDP: %v", err))
					}
				}
			}
		}
	}(l, result)
	return
}

func (s *Server) serverRunNewMessage(message msg.Message, remoteAddr *net.UDPAddr) (wm msg.Message, wra *net.UDPAddr, w bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	remoteAddrS := remoteAddr.String()
	split := strings.Split(remoteAddrS, ":")
	port, _ := strconv.Atoi(split[1])

	switch m := message.(type) {
	case *msg.UDPNewConnectRequestMessage:
		if cSid, ok := s.registerNameMap[m.Name]; ok {
			if writeChanMessage, ok := s.writeChanMap[cSid]; ok {
				s.p2pMap[m.Cid] = remoteAddr
				writeChanMessage.writeChan <- &msg.UDPNewConnectResponseMessage{Cid: m.Cid, CIp: split[0], CPort: uint16(port), Name: m.Name}
				log.Info(m.Cid, 0, remoteAddrS)
			} else {
				log.Error(m.Cid, 0, fmt.Errorf("writeChanMap no find [%d]", cSid))
			}
		} else {
			log.Error(m.Cid, 0, fmt.Errorf("registerNameMap no find [%s]", m.Name))
		}
	case *msg.UDPNewConnectResultRequestMessage:
		wm = &msg.UDPNewConnectResultResponseMessage{Cid: m.Cid, Sid: m.Sid, SIp: split[0], SPort: uint16(port)}
		wra = s.p2pMap[m.Cid]
		w = true
		log.Info(m.Cid, m.Sid, remoteAddrS)
	default:
		log.Error(0, 0, fmt.Errorf("serverRunNewMessage: %v", message.ToByteBuf()))
	}
	return
}

func (s *Server) NewConnect(int32, int32, chan msg.Message) {}

func (s *Server) ConnectSuccess(cId int32, sId int32, writeChan chan msg.Message) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.writeChanMap[sId] = &WriteChanMessage{cId, sId, writeChan}
}

func (s *Server) CloseConnect(cId int32, sId int32) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 通知客户端已失效
	if names, ok := s.registerIdMap[sId]; ok {
		accessIdsCache := make(map[int32][]string)
		for _, name := range names {
			if accessIds, ok := s.accessNameMap[name]; ok {
				for accessId := range accessIds {
					accessIdsCache[accessId] = append(accessIdsCache[accessId], name)
					delete(s.accessIdMap[accessId], name)
				}
			}
			delete(s.registerNameMap, name)
			delete(s.accessNameMap, name)
		}
		for accessSid, cNames := range accessIdsCache {
			if writeChanMessage, ok := s.writeChanMap[accessSid]; ok {
				s.hasErrorAccessMap[accessSid] = 1
				writeChanMessage.writeChan <- &msg.UDPAccessCloseResponseMessage{CloseNames: cNames}
				log.Info(writeChanMessage.cId, writeChanMessage.sId, fmt.Sprintf("Sub Access: %v", cNames))
			}
		}
		delete(s.registerIdMap, sId)
		log.Info(cId, sId, fmt.Sprintf("Sub Register: %v", names))
	}
	if names, ok := s.accessIdMap[sId]; ok {
		for name := range names {
			if accessSids, ok := s.accessNameMap[name]; ok {
				log.Info(cId, sId, fmt.Sprintf("Sub Access: %v", names))
				delete(accessSids, sId)
			}
		}
		delete(s.accessIdMap, sId)
	}
	delete(s.writeChanMap, sId)
}

func (s *Server) Handle(cId int32, sId int32, ip string, port uint16, message *io.Message) (handle bool) {
	switch message.Message.(type) {
	case *msg.UDPRegisterRequestMessage:
	case *msg.UDPAccessRequestMessage:
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

	switch m := message.Message.(type) {
	case *msg.UDPRegisterRequestMessage:
		// 服务端, 有一个失败的直接拒绝
		for _, name := range m.Names {
			if _, nameOk := s.registerNameMap[name]; nameOk {
				writeChanMessage.writeChan <- &msg.UDPRegisterResponseMessage{Status: msg.UDPRegisterError, Msg: "exist " + name}
				return
			}
		}
		s.registerIdMap[sId] = m.Names
		for _, name := range m.Names {
			s.registerNameMap[name] = sId
		}
		writeChanMessage.writeChan <- &msg.UDPRegisterResponseMessage{Status: msg.UDPRegisterSuccess, ServerUDPPort: udpPort}
		// 通知所有失败的客户端来刷新
		for clientId := range s.hasErrorAccessMap {
			if c, ok := s.writeChanMap[clientId]; ok {
				log.Trace(clientId, 0, "flush names")
				c.writeChan <- &msg.UDPAccessFlushResponseMessage{}
			}
			delete(s.hasErrorAccessMap, clientId)
		}
		log.Info(cId, sId, fmt.Sprintf("%s: %v", "register", m.Names))
	case *msg.UDPAccessRequestMessage:
		// 客户端, 返回失败的
		waitName := make([]string, 0)
		successName := make([]string, 0)
		errorName := make([]string, 0)
		// 避免重复提交
		if tempMap, ok := s.accessIdMap[sId]; ok {
			for _, name := range m.Names {
				if _, ok := tempMap[name]; !ok {
					waitName = append(waitName, name)
				}
			}
		} else {
			waitName = m.Names
		}

		for _, name := range waitName {
			if _, nameOk := s.registerNameMap[name]; nameOk {
				successName = append(successName, name)
			} else {
				errorName = append(errorName, name)
			}
		}

		for _, name := range successName {
			csId := s.registerNameMap[name]
			if idMap, ok := s.accessIdMap[sId]; ok {
				idMap[name] = csId
			} else {
				s.accessIdMap[sId] = make(map[string]int32)
				s.accessIdMap[sId][name] = csId
			}
			if nameMap, ok := s.accessNameMap[name]; ok {
				nameMap[sId] = csId
			} else {
				s.accessNameMap[name] = make(map[int32]int32)
				s.accessNameMap[name][sId] = csId
			}
		}
		if len(errorName) > 0 {
			s.hasErrorAccessMap[sId] = 1
		}
		writeChanMessage.writeChan <- &msg.UDPAccessResponseMessage{ServerUDPPort: udpPort, SuccessNames: successName}
		log.Info(cId, sId, fmt.Sprintf("%s: success[%v], error[%v]", "access", successName, errorName))
	default:
		handle = false
	}
	return
}
