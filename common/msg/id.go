package msg

import (
	"p2p-go/common/utils"
)

type IdRequestMessage struct {
	Id int32
}

func (m *IdRequestMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		return &IdRequestMessage{id}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *IdRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = IdRequest
	copy(buf[1:], utils.Id2Buf(m.Id))
	return
}

type IdResponseMessage struct {
	Id int32
}

func (m *IdResponseMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		return &IdResponseMessage{id}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *IdResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = IdResponse
	copy(buf[1:], utils.Id2Buf(m.Id))
	return
}
