package transfer_tcp

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

type Client struct {
	writeChan          chan msg.Message
	registerReloadUUID time.Time
	serverTCPAddr      string
}

func NewClient(writeChan chan msg.Message) *Client {
	return &Client{
		writeChan,
		time.Now(),
		config.ServerIp + ":" + strconv.Itoa(config.ServerPort),
	}
}

func (c *Client) ChangeStatus(cId int32, sId int32, status byte) {
	tunLock.Lock()
	defer tunLock.Unlock()
	if status == handle.Success {
		log.Trace(cId, sId, "send register request")
		c.writeChan <- &msg.TCPTunRegisterRequestMessage{Buf: tunToBuf(noRegister)}
	} else if status == handle.Error {
		log.Info(cId, sId, "change status register error")
		registerSuccess = make(map[int32]tun)
		noRegister = make(map[int32]tun)
		for k, v := range tuns {
			noRegister[k] = v
		}
	}
}

func (c *Client) Handle(cId int32, sId int32, message *io.Message) (handle bool) {
	switch message.Message.(type) {
	case *msg.TCPTunRegisterResponseMessage:
	case *msg.TCPTunNewConnectRequestMessage:
	default:
		return
	}

	handle = true
	switch m := message.Message.(type) {
	case *msg.TCPTunRegisterResponseMessage:
		tunLock.Lock()
		defer tunLock.Unlock()
		for _, id := range m.Ids {
			if t, ok := noRegister[id]; ok {
				registerSuccess[id] = t
				delete(noRegister, id)
				log.Info(cId, sId, fmt.Sprintf("tun register success [%s:%d=%s:%d]", t.remoteAddr, t.remotePort, t.localAddr, t.localPort))
			} else {
				log.Error(cId, sId, fmt.Errorf("tun reggister success but no find [%d]", id))
			}
		}
	case *msg.TCPTunNewConnectRequestMessage:
		if tun, ok := registerSuccess[m.Id]; ok {
			log.Trace(0, m.Sid, "new tcp tun connect")
			serverAddr, err := net.ResolveTCPAddr("tcp4", c.serverTCPAddr)
			if err != nil {
				log.Error(cId, sId, fmt.Errorf("resolve tcp addr error, sid: [%d], message: %v", m.Sid, err))
				c.writeChan <- &msg.TCPTunNewConnectResponseMessage{Sid: m.Sid, Message: err.Error()}
				break
			}
			serverConn, err := net.DialTCP("tcp4", nil, serverAddr)
			if err != nil {
				log.Error(cId, sId, fmt.Errorf("dial tcp error, sid: [%d], message: %v", m.Sid, err))
				c.writeChan <- &msg.TCPTunNewConnectResponseMessage{Sid: m.Sid, Message: err.Error()}
				break
			}
			serverTCP := io.NewTCP(serverConn)
			serverTCP.Tid = m.Sid

			addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", tun.localAddr, tun.localPort))
			if err != nil {
				log.Error(cId, sId, fmt.Errorf("resolve tcp addr error, sid: [%d], message: %v", m.Sid, err))
				if _, err = serverTCP.WriteMessage(&msg.TCPTunNewConnectResponseMessage{Sid: m.Sid, Message: err.Error()}); err != nil {
					c.writeChan <- &msg.TCPTunNewConnectResponseMessage{Sid: m.Sid, Message: err.Error()}
				}
				_ = serverTCP.Close()
				break
			}
			appConn, err := net.DialTCP("tcp4", nil, addr)
			if err != nil {
				log.Error(cId, sId, fmt.Errorf("dial tcp error, sid: [%d], message: %v", m.Sid, err))
				if _, err = serverTCP.WriteMessage(&msg.TCPTunNewConnectResponseMessage{Sid: m.Sid, Message: err.Error()}); err != nil {
					c.writeChan <- &msg.TCPTunNewConnectResponseMessage{Sid: m.Sid, Message: err.Error()}
				}
				_ = serverTCP.Close()
				break
			}
			appTCP := io.NewTCP(appConn)
			appTCP.Tid = m.Sid
			if _, err = serverTCP.WriteMessage(&msg.TCPTunNewConnectResponseMessage{Cid: appTCP.Id, Sid: m.Sid}); err != nil {
				_ = serverTCP.Close()
				_ = appConn.Close()
			}
			log.Info(cId, sId, fmt.Sprintf("new tcp tun success, serverTCP[%d], appTCP[%d], sId[%d]", serverTCP.Id, appTCP.Id, m.Sid))

			go goTimeOut(serverTCP)
			go goTimeOut(appTCP)
			go transferData(serverTCP, appTCP)
			go transferData(appTCP, serverTCP)
		} else {
			c.writeChan <- &msg.TCPTunNewConnectResponseMessage{Sid: m.Sid, Message: "no register success"}
			log.Error(cId, sId, fmt.Errorf("tun new connect but registerSuccess no find [%d], sId: [%d]", m.Id, m.Sid))
		}
	default:
		handle = false
	}
	return
}
