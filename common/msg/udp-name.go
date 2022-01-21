package msg

import "p2p-go/common/utils"

const (
	_ byte = iota
	UDPRegisterError
	UDPRegisterSuccess
)

type UDPRegisterRequestMessage struct {
	Names []string
}

func (m *UDPRegisterRequestMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	i := 1
	for i < len(srcBuf) {
		endIndex := i + 1 + int(srcBuf[i])
		names = append(names, string(srcBuf[i+1:endIndex]))
		i = endIndex
	}
	return &UDPRegisterRequestMessage{names}
}
func (m *UDPRegisterRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = UDPRegisterRequest
	for _, name := range m.Names {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type UDPRegisterResponseMessage struct {
	Status        byte
	ServerUDPPort uint16
	Msg           string
}

func (m *UDPRegisterResponseMessage) toMessage(srcBuf []byte) Message {
	msg := ""
	if len(srcBuf) > 4 {
		msg = string(srcBuf[4:])
	}
	if port, err := utils.Buf2Port(srcBuf[2:]); err == nil {
		return &UDPRegisterResponseMessage{srcBuf[1], port, msg}
	}
	return &ErrorMessage{srcBuf}
}
func (m *UDPRegisterResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 4)
	buf[0] = UDPRegisterResponse
	buf[1] = m.Status
	copy(buf[2:], utils.Port2Buf(m.ServerUDPPort))
	buf = append(buf, []byte(m.Msg)...)
	return
}

type UDPAccessRequestMessage struct {
	Names []string
}

func (m *UDPAccessRequestMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	i := 1
	for i < len(srcBuf) {
		endIndex := i + 1 + int(srcBuf[i])
		names = append(names, string(srcBuf[i+1:endIndex]))
		i = endIndex
	}
	return &UDPAccessRequestMessage{names}
}
func (m *UDPAccessRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = UDPAccessRequest
	for _, name := range m.Names {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type UDPAccessResponseMessage struct {
	ServerUDPPort uint16
	SuccessNames  []string
}

func (m *UDPAccessResponseMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	if len(srcBuf) > 3 {
		i := 3
		for i < len(srcBuf) {
			endIndex := i + 1 + int(srcBuf[i])
			names = append(names, string(srcBuf[i+1:endIndex]))
			i = endIndex
		}
	}
	if port, err := utils.Buf2Port(srcBuf[1:]); err == nil {
		return &UDPAccessResponseMessage{port, names}
	}
	return &ErrorMessage{srcBuf}
}
func (m *UDPAccessResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 3)
	buf[0] = UDPAccessResponse
	copy(buf[1:], utils.Port2Buf(m.ServerUDPPort))
	for _, name := range m.SuccessNames {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type UDPAccessCloseResponseMessage struct {
	CloseNames []string
}

func (m *UDPAccessCloseResponseMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	if len(srcBuf) > 1 {
		i := 1
		for i < len(srcBuf) {
			endIndex := i + 1 + int(srcBuf[i])
			names = append(names, string(srcBuf[i+1:endIndex]))
			i = endIndex
		}
	}
	return &UDPAccessCloseResponseMessage{names}
}
func (m *UDPAccessCloseResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = UDPAccessCloseResponse
	for _, name := range m.CloseNames {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type UDPAccessFlushResponseMessage struct{}

func (m *UDPAccessFlushResponseMessage) toMessage([]byte) Message {
	return &UDPAccessFlushResponseMessage{}
}
func (m *UDPAccessFlushResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = UDPAccessFlushResponse
	return
}
