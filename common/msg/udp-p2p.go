package msg

import "p2p-go/common/utils"

type UDPNewConnectRequestMessage struct {
	Cid  int32
	Name string
}

func (m *UDPNewConnectRequestMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		return &UDPNewConnectRequestMessage{id, string(srcBuf[5:])}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *UDPNewConnectRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 5)
	buf[0] = UDPNewConnectRequest
	copy(buf[1:], utils.Id2Buf(m.Cid))
	buf = append(buf, []byte(m.Name)...)
	return
}

type UDPNewConnectResponseMessage struct {
	Cid   int32
	CIp   string
	CPort uint16
	Name  string
}

func (m *UDPNewConnectResponseMessage) toMessage(srcBuf []byte) Message {
	if id, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if ip, err := utils.Buf2Ip(srcBuf[5:]); err == nil {
			if port, err := utils.Buf2Port(srcBuf[9:]); err == nil {
				return &UDPNewConnectResponseMessage{id, ip, port, string(srcBuf[11:])}
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
func (m *UDPNewConnectResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 11)
	buf[0] = UDPNewConnectResponse
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

type UDPNewConnectResultRequestMessage struct {
	Cid int32
	Sid int32
}

func (m *UDPNewConnectResultRequestMessage) toMessage(srcBuf []byte) Message {
	if cId, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if sId, err := utils.Buf2Id(srcBuf[5:]); err == nil {
			return &UDPNewConnectResultRequestMessage{cId, sId}
		} else {
			return &ErrorMessage{srcBuf}
		}
	} else {
		return &ErrorMessage{srcBuf}
	}
}
func (m *UDPNewConnectResultRequestMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 9)
	buf[0] = UDPNewConnectResultRequest
	copy(buf[1:], utils.Id2Buf(m.Cid))
	copy(buf[5:], utils.Id2Buf(m.Sid))
	return
}

type UDPNewConnectResultResponseMessage struct {
	Cid   int32
	Sid   int32
	SIp   string
	SPort uint16
}

func (m *UDPNewConnectResultResponseMessage) toMessage(srcBuf []byte) Message {
	if cId, err := utils.Buf2Id(srcBuf[1:]); err == nil {
		if sId, err := utils.Buf2Id(srcBuf[5:]); err == nil {
			if ip, err := utils.Buf2Ip(srcBuf[9:]); err == nil {
				if port, err := utils.Buf2Port(srcBuf[13:]); err == nil {
					return &UDPNewConnectResultResponseMessage{cId, sId, ip, port}
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
func (m *UDPNewConnectResultResponseMessage) ToByteBuf() (buf []byte) {
	buf = make([]byte, 15)
	buf[0] = UDPNewConnectResultResponse
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
