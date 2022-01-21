package msg

type IgnoreMessage struct {
}

func (m *IgnoreMessage) toMessage([]byte) Message {
	return &IgnoreMessage{}
}
func (m *IgnoreMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1)
	buf[0] = Ignore
	return
}
