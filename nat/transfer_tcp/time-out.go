package transfer_tcp

import (
	sio "io"
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

func transferData(dest *io.TCP, src *io.TCP) {
	defer func() {
		_ = dest.Close()
		_ = src.Close()
	}()

	buf := make([]byte, 64*1024)
	for {
		readLen, err := src.Read(buf)
		if src.TimeOut || dest.TimeOut {
			break
		}
		if err != nil {
			if err != sio.EOF {
				log.Error(0, 0, err)
			}
			break
		}
		_, err = dest.Write(buf[:readLen])
		if err != nil {
			if err != sio.EOF {
				log.Error(0, 0, err)
			}
			break
		}
		src.LastTransferTime = time.Now()
		dest.LastTransferTime = time.Now()
	}
}
