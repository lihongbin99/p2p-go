package p2p_udp

import (
	"p2p-go/common/io"
	"time"
)

var (
	timeOut = 5 * time.Minute
)

func goTimeOut(udp *io.UDP) {
	defer func() {
		udp.TimeOut = true
		_ = udp.Close()
	}()
	time.Sleep(timeOut)
	for {
		if time.Now().Sub(udp.LastTransferTime) > timeOut {
			break
		}
		time.Sleep(udp.LastTransferTime.Add(timeOut).Sub(time.Now()))
	}
}
