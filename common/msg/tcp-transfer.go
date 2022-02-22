package msg

import "p2p-go/common/utils"

type TCPTransferRequestMessage struct {
	Sid  int32
	Name string
}

func (m *TCPTransferRequestMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		return &TCPTransferRequestMessage{id, string(srcBuf[5:])}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *TCPTransferRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = TCPTransferRequest
	copy(buf[1:], utils.Id2Buf(m.Sid))
	buf = append(buf, []byte(m.Name)...)
	return
}

type TCPTransferResponseMessage struct {
	Sid     int32
	Message string
}

func (m *TCPTransferResponseMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if len(srcBuf) > 5 {
			return &TCPTransferResponseMessage{id, string(srcBuf[5:])}
		} else {
			return &TCPTransferResponseMessage{id, ""}
		}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *TCPTransferResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = TCPTransferResponse
	copy(buf[1:], utils.Id2Buf(m.Sid))
	if m.Message != "" {
		buf = append(buf, []byte(m.Message)...)
	}
	return
}
