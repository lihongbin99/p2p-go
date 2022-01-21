package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func Id2Buf(id int32) (buf []byte) {
	buf = make([]byte, 4)
	buf[0] = byte(id >> 24)
	buf[1] = byte(id >> 18)
	buf[2] = byte(id >> 8)
	buf[3] = byte(id)
	return
}

func Buf2Id(buf []byte) (id int32, err error) {
	if len(buf) >= 4 {
		id = int32(buf[0])<<24 + int32(buf[1])<<16 + int32(buf[2])<<8 + int32(buf[3])
	} else {
		err = fmt.Errorf("len(buf) = %d", len(buf))
	}
	return
}

func Ip2Buf(ip string) (buf []byte, err error) {
	buf = make([]byte, 4)
	sa := strings.Split(ip, ".")
	if len(sa) >= 4 {
		for i := 0; i < 4; i++ {
			num, atoiErr := strconv.Atoi(sa[i])
			if atoiErr != nil {
				err = atoiErr
				return
			}
			buf[i] = byte(num)
		}
	} else {
		err = fmt.Errorf("len(ip.) = %d", len(buf))
	}
	return
}

func Buf2Ip(buf []byte) (ip string, err error) {
	if len(buf) >= 4 {
		ip = strconv.Itoa(int(buf[0])) + "." + strconv.Itoa(int(buf[1])) + "." + strconv.Itoa(int(buf[2])) + "." + strconv.Itoa(int(buf[3]))
	} else {
		err = fmt.Errorf("len(buf) = %d", len(buf))
	}
	return
}

func Port2Buf(port uint16) (buf []byte) {
	buf = make([]byte, 2)
	buf[0] = byte(port >> 8)
	buf[1] = byte(port)
	return
}

func Buf2Port(buf []byte) (port uint16, err error) {
	if len(buf) >= 2 {
		port = uint16(buf[0])<<8 + uint16(buf[1])
	} else {
		err = fmt.Errorf("len(buf) = %d", len(buf))
	}
	return
}
