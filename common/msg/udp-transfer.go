package msg

import "p2p-go/common/utils"

type UDPTransferRequestMessage struct {
	Sid  int32
	Name string
}

func (m *UDPTransferRequestMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		return &UDPTransferRequestMessage{id, string(srcBuf[5:])}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *UDPTransferRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = UDPTransferRequest
	copy(buf[1:], utils.Id2Buf(m.Sid))
	buf = append(buf, []byte(m.Name)...)
	return
}

type UDPTransferResponseMessage struct {
	Sid     int32
	Message string
}

func (m *UDPTransferResponseMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if len(srcBuf) > 5 {
			return &UDPTransferResponseMessage{id, string(srcBuf[5:])}
		} else {
			return &UDPTransferResponseMessage{id, ""}
		}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *UDPTransferResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = UDPTransferResponse
	copy(buf[1:], utils.Id2Buf(m.Sid))
	if m.Message != "" {
		buf = append(buf, []byte(m.Message)...)
	}
	return
}
