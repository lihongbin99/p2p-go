package main

import (
	"fmt"
	"net"
	logger "p2p-go/common/log"
	"time"
)

func main() {
	log := logger.Log{From: "provide"}
	addr, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:13522")
	if err != nil {
		log.Fatal(err)
	}

	listen, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 13)

	tcp, err := listen.AcceptTCP()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		readLength, err := tcp.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(0, 0, fmt.Sprintf("Test-%d: %s", i, string(buf[:readLength])))
	}

	buf = make([]byte, 64*1024)

	// 下载测速
	maxRead := 0
	readLength, err := tcp.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(0, 0, "Start")
	maxRead += readLength
	startTime := time.Now()
	endTime := time.Now()

	for {
		_ = tcp.SetReadDeadline(time.Now().Add(3 * time.Second))
		readLength, err = tcp.Read(buf)
		if err != nil {
			log.Error(0, 0, err)
			break
		}
		_ = tcp.SetReadDeadline(time.Time{})
		maxRead += readLength
		endTime = time.Now()
	}

	v := endTime.Sub(startTime) / time.Second
	log.Info(0, 0, fmt.Sprintf("maxDownload: %v", maxRead))
	log.Info(0, 0, fmt.Sprintf("v: %v", int(v)))
	log.Info(0, 0, fmt.Sprintf("Download: %dMB/s", maxRead/1024/1024/int(v)))

	// 上传测速
	time.Sleep(1 * time.Second)

	maxWrite := 0
	c := true
	ticker := time.NewTicker(3 * time.Second)
	for c {
		select {
		case _ = <-ticker.C:
			c = false
			break
		default:
			writeLength, err := tcp.Write(buf)
			if err != nil {
				log.Fatal(err)
			}
			maxWrite += writeLength
			//time.Sleep(10 * time.Millisecond)
		}
	}

	log.Info(0, 0, fmt.Sprintf("maxUpload: %v", maxWrite))
	log.Info(0, 0, fmt.Sprintf("Upload: %dMB/s", maxWrite/1024/1024/3))
}
