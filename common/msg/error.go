package msg

type ErrorMessage struct {
	Buf []byte
}

func (m *ErrorMessage) toMessage(srcBuf []byte) Message {
	buf := make([]byte, len(srcBuf))
	copy(buf, srcBuf)
	return &ErrorMessage{buf}
}
func (m *ErrorMessage) ToByteBuf() []byte {
	buf := make([]byte, len(m.Buf))
	copy(buf, m.Buf)
	return buf[:len(m.Buf)]
}
