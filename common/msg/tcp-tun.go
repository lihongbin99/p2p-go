package msg

import "p2p-go/common/utils"

type TCPTunRegisterRequestMessage struct {
	Buf []byte
}

func (m *TCPTunRegisterRequestMessage) toMessage(srcBuf []byte) Message {
	return &TCPTunRegisterRequestMessage{
		Buf: srcBuf[1:],
	}
}
func (m *TCPTunRegisterRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = TCPTunRegisterRequest
	if len(m.Buf) > 0 {
		buf = append(buf, m.Buf...)
	}
	return
}

type TCPTunRegisterResponseMessage struct {
	Ids []int32
}

func (m *TCPTunRegisterResponseMessage) toMessage(srcBuf []byte) Message {
	ids := make([]int32, 0)
	buf := srcBuf[1:]
	if len(buf) >= 4 {
		id, _ := utils.Buf2Id(buf[0:4])
		ids = append(ids, id)
		buf = buf[4:]
	}
	return &TCPTunRegisterResponseMessage{Ids: ids}
}
func (m *TCPTunRegisterResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = TCPTunRegisterResponse
	for _, id := range m.Ids {
		buf = append(buf, utils.Id2Buf(id)...)
	}
	return
}

type TCPTunNewConnectRequestMessage struct {
	Id  int32
	Sid int32
}

func (m *TCPTunNewConnectRequestMessage) toMessage(srcBuf []byte) Message {
	id, _ := utils.Buf2Id(srcBuf[1:5])
	sid, _ := utils.Buf2Id(srcBuf[5:9])
	return &TCPTunNewConnectRequestMessage{Id: id, Sid: sid}
}
func (m *TCPTunNewConnectRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 9)
	buf[0] = TCPTunNewConnectRequest
	copy(buf[1:], utils.Id2Buf(m.Id))
	copy(buf[5:], utils.Id2Buf(m.Sid))
	return
}

type TCPTunNewConnectResponseMessage struct {
	Cid     int32
	Sid     int32
	Message string
}

func (m *TCPTunNewConnectResponseMessage) toMessage(srcBuf []byte) Message {
	cid, _ := utils.Buf2Id(srcBuf[1:5])
	sid, _ := utils.Buf2Id(srcBuf[5:9])
	message := ""
	if len(srcBuf) > 9 {
		message = string(srcBuf[9:])
	}
	return &TCPTunNewConnectResponseMessage{Cid: cid, Sid: sid, Message: message}
}
func (m *TCPTunNewConnectResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 9)
	buf[0] = TCPTunNewConnectResponse
	copy(buf[1:], utils.Id2Buf(m.Cid))
	copy(buf[5:], utils.Id2Buf(m.Sid))
	if len(m.Message) > 0 {
		buf = append(buf, []byte(m.Message)...)
	}
	return
}
