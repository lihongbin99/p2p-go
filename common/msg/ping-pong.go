package msg

import "time"

type PingMessage struct {
	Date time.Time
}

func (m *PingMessage) toMessage(srcBuf []byte) Message {
	date, err := time.Parse(layout, string(srcBuf[1:1+len(layout)]))
	if err != nil {
		return nil
	}
	return &PingMessage{date}
}
func (m *PingMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1+len(layout))
	buf[0] = Ping
	copy(buf[1:], m.Date.Format(layout))
	return
}

type PongMessage struct {
	Date time.Time
}

func (m *PongMessage) toMessage(srcBuf []byte) Message {
	date, err := time.Parse(layout, string(srcBuf[1:1+len(layout)]))
	if err != nil {
		return nil
	}
	return &PongMessage{date}
}
func (m *PongMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 1+len(layout))
	buf[0] = Pong
	copy(buf[1:], m.Date.Format(layout))
	return
}
