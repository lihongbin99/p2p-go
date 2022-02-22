package p2p_tcp

import (
	"p2p-go/common/io"
	"time"
)

var (
	timeOut = 5 * time.Minute
)

func goTimeOut(tcp *io.TCP) {
	time.Sleep(timeOut)
	for {
		if time.Now().Sub(tcp.LastTransferTime) > timeOut {
			break
		}
		time.Sleep(tcp.LastTransferTime.Add(timeOut).Sub(time.Now()))
	}
	tcp.TimeOut = true
	_ = tcp.Close()
}
