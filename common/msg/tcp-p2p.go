package msg

import "p2p-go/common/utils"

type TCPNewConnectRequestMessage struct {
	Cid  int32
	Name string
}

func (m *TCPNewConnectRequestMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		return &TCPNewConnectRequestMessage{id, string(srcBuf[5:])}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *TCPNewConnectRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = TCPNewConnectRequest
	copy(buf[1:], utils.Id2Buf(m.Cid))
	buf = append(buf, []byte(m.Name)...)
	return
}

type TCPNewConnectResponseMessage struct {
	Cid   int32
	CIp   string
	CPort uint16
	Name  string
}

func (m *TCPNewConnectResponseMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if ip, err := utils.Buf2Ip(srcBuf[5:]); err == nil {
			if port, err := utils.Buf2Port(srcBuf[9:]); err == nil {
				return &TCPNewConnectResponseMessage{id, ip, port, string(srcBuf[11:])}
			} else {
				return &ErrorMessage{srcBuf}
			}
		} else {
			return &ErrorMessage{srcBuf}
		}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *TCPNewConnectResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 11)
	buf[0] = TCPNewConnectResponse
	copy(buf[1:], utils.Id2Buf(m.Cid))
	if ipb, err := utils.Ip2Buf(m.CIp); err == nil {
		copy(buf[5:], ipb)
	} else {
		copy(buf[5:], []byte{0, 0, 0, 0})
	}
	copy(buf[9:], utils.Port2Buf(m.CPort))
	buf = append(buf, []byte(m.Name)...)
	return
}

type TCPNewConnectResultRequestMessage struct {
	Cid int32
	Sid int32
}

func (m *TCPNewConnectResultRequestMessage) toMessage(srcBuf []byte) Message {
	if cId, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if sId, err := utils.Buf2Id(srcBuf[5:]); err == nil {
			return &TCPNewConnectResultRequestMessage{cId, sId}
		} else {
			return &ErrorMessage{srcBuf}
		}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *TCPNewConnectResultRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 9)
	buf[0] = TCPNewConnectResultRequest
	copy(buf[1:], utils.Id2Buf(m.Cid))
	copy(buf[5:], utils.Id2Buf(m.Sid))
	return
}

type TCPNewConnectResultResponseMessage struct {
	Cid   int32
	Sid   int32
	SIp   string
	SPort uint16
}

func (m *TCPNewConnectResultResponseMessage) toMessage(srcBuf []byte) Message {
	if cId, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if sId, err := utils.Buf2Id(srcBuf[5:]); err == nil {
			if ip, err := utils.Buf2Ip(srcBuf[9:]); err == nil {
				if port, err := utils.Buf2Port(srcBuf[13:]); err == nil {
					return &TCPNewConnectResultResponseMessage{cId, sId, ip, port}
				} else {
					return &ErrorMessage{srcBuf}
				}
			} else {
				return &ErrorMessage{srcBuf}
			}
		} else {
			return &ErrorMessage{srcBuf}
		}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *TCPNewConnectResultResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 15)
	buf[0] = TCPNewConnectResultResponse
	copy(buf[1:], utils.Id2Buf(m.Cid))
	copy(buf[5:], utils.Id2Buf(m.Sid))
	if ipb, err := utils.Ip2Buf(m.SIp); err == nil {
		copy(buf[9:], ipb)
	} else {
		copy(buf[9:], []byte{0, 0, 0, 0})
	}
	copy(buf[13:], utils.Port2Buf(m.SPort))
	return
}
