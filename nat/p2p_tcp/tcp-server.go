package p2p_tcp

import (
	"fmt"
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
	writeChanMap      map[int32]*WriteChanMessage
	registerIdMap     map[int32][]string         // Map<sId, List<sName> >
	registerNameMap   map[string]int32           // Map<sName, sId>
	accessIdMap       map[int32]map[string]int32 // Map<cId, Map<cName, sId> >
	accessNameMap     map[string]map[int32]int32 // Map<cName, Map<cId, sId> >
	hasErrorAccessMap map[int32]interface{}      // Map<cId, ant>

	p2pMap map[int32]*WriteChanMessage // Map<cId, cRemoteAddr>
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
		make(map[int32]*WriteChanMessage),
		sync.Mutex{},
	}
	return
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
				writeChanMessage.writeChan <- &msg.TCPAccessCloseResponseMessage{CloseNames: cNames}
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
	case *msg.TCPRegisterRequestMessage:
	case *msg.TCPAccessRequestMessage:
	case *msg.TCPNewConnectRequestMessage:
	case *msg.TCPNewConnectResultRequestMessage:
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
	case *msg.TCPRegisterRequestMessage:
		// 服务端, 有一个失败的直接拒绝
		for _, name := range m.Names {
			if _, nameOk := s.registerNameMap[name]; nameOk {
				writeChanMessage.writeChan <- &msg.TCPRegisterResponseMessage{Status: msg.TCPRegisterError, Msg: "exist " + name}
				return
			}
		}
		s.registerIdMap[sId] = m.Names
		for _, name := range m.Names {
			s.registerNameMap[name] = sId
		}
		writeChanMessage.writeChan <- &msg.TCPRegisterResponseMessage{Status: msg.TCPRegisterSuccess}
		// 通知所有失败的客户端来刷新
		for clientId := range s.hasErrorAccessMap {
			if c, ok := s.writeChanMap[clientId]; ok {
				log.Trace(clientId, 0, "flush names")
				c.writeChan <- &msg.TCPAccessFlushResponseMessage{}
			}
			delete(s.hasErrorAccessMap, clientId)
		}
		log.Info(cId, sId, fmt.Sprintf("register: %v", m.Names))
	case *msg.TCPAccessRequestMessage:
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
		writeChanMessage.writeChan <- &msg.TCPAccessResponseMessage{SuccessNames: successName}
		log.Info(cId, sId, fmt.Sprintf("%s: success%v, error%v", "access", successName, errorName))
	case *msg.TCPNewConnectRequestMessage:
		if cSid, ok := s.registerNameMap[m.Name]; ok {
			if csWriteChanMessage, ok := s.writeChanMap[cSid]; ok {
				s.p2pMap[m.Cid] = writeChanMessage
				csWriteChanMessage.writeChan <- &msg.TCPNewConnectResponseMessage{Cid: m.Cid, CIp: ip, CPort: port, Name: m.Name}
				log.Info(m.Cid, 0, fmt.Sprintf("%s:%d", ip, port))
			} else {
				log.Error(m.Cid, 0, fmt.Errorf("writeChanMap no find [%d]", cSid))
			}
		} else {
			log.Error(m.Cid, 0, fmt.Errorf("registerNameMap no find [%s]", m.Name))
		}
	case *msg.TCPNewConnectResultRequestMessage:
		s.p2pMap[m.Cid].writeChan <- &msg.TCPNewConnectResultResponseMessage{Cid: m.Cid, Sid: m.Sid, SIp: ip, SPort: port}
		log.Info(m.Cid, m.Sid, fmt.Sprintf("%s:%d", ip, port))
	default:
		handle = false
	}
	return
}
