package msg

const (
	_ byte = iota
	TCPRegisterError
	TCPRegisterSuccess
)

type TCPRegisterRequestMessage struct {
	Names []string
}

func (m *TCPRegisterRequestMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	i := 1
	for i < len(srcBuf) {
		endIndex := i + 1 + int(srcBuf[i])
		names = append(names, string(srcBuf[i+1:endIndex]))
		i = endIndex
	}
	return &TCPRegisterRequestMessage{names}
}
func (m *TCPRegisterRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = TCPRegisterRequest
	for _, name := range m.Names {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type TCPRegisterResponseMessage struct {
	Status byte
	Msg    string
}

func (m *TCPRegisterResponseMessage) toMessage(srcBuf []byte) Message {
	msg := ""
	if len(srcBuf) > 2 {
		msg = string(srcBuf[2:])
	}
	return &TCPRegisterResponseMessage{srcBuf[1], msg}
}
func (m *TCPRegisterResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 2)
	buf[0] = TCPRegisterResponse
	buf[1] = m.Status
	buf = append(buf, []byte(m.Msg)...)
	return
}

type TCPAccessRequestMessage struct {
	Names []string
}

func (m *TCPAccessRequestMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	i := 1
	for i < len(srcBuf) {
		endIndex := i + 1 + int(srcBuf[i])
		names = append(names, string(srcBuf[i+1:endIndex]))
		i = endIndex
	}
	return &TCPAccessRequestMessage{names}
}
func (m *TCPAccessRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = TCPAccessRequest
	for _, name := range m.Names {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type TCPAccessResponseMessage struct {
	SuccessNames []string
}

func (m *TCPAccessResponseMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	if len(srcBuf) > 1 {
		i := 1
		for i < len(srcBuf) {
			endIndex := i + 1 + int(srcBuf[i])
			names = append(names, string(srcBuf[i+1:endIndex]))
			i = endIndex
		}
	}
	return &TCPAccessResponseMessage{names}
}
func (m *TCPAccessResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = TCPAccessResponse
	for _, name := range m.SuccessNames {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type TCPAccessCloseResponseMessage struct {
	CloseNames []string
}

func (m *TCPAccessCloseResponseMessage) toMessage(srcBuf []byte) Message {
	names := make([]string, 0)
	if len(srcBuf) > 1 {
		i := 1
		for i < len(srcBuf) {
			endIndex := i + 1 + int(srcBuf[i])
			names = append(names, string(srcBuf[i+1:endIndex]))
			i = endIndex
		}
	}
	return &TCPAccessCloseResponseMessage{names}
}
func (m *TCPAccessCloseResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = TCPAccessCloseResponse
	for _, name := range m.CloseNames {
		buf = append(buf, byte(len(name)))
		buf = append(buf, []byte(name)...)
	}
	return
}

type TCPAccessFlushResponseMessage struct{}

func (m *TCPAccessFlushResponseMessage) toMessage([]byte) Message {
	return &TCPAccessFlushResponseMessage{}
}
func (m *TCPAccessFlushResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = TCPAccessFlushResponse
	return
}
