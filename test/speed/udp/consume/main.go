package main

import (
	"fmt"
	"net"
	logger "p2p-go/common/log"
	"strconv"
	"time"
)

func main() {
	log := logger.Log{From: "consume"}
	addr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:13521")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		_, err = conn.Write([]byte("Hello World-" + strconv.Itoa(i)))
		if err != nil {
			log.Fatal(err)
		}
	}

	// 上传测速
	time.Sleep(1 * time.Second)

	buf := make([]byte, 1500-36)

	maxWrite := 0
	c := true
	ticker := time.NewTicker(3 * time.Second)
	for c {
		select {
		case _ = <-ticker.C:
			c = false
			break
		default:
			writeLength, err := conn.Write(buf)
			if err != nil {
				log.Fatal(err)
			}
			maxWrite += writeLength
			//time.Sleep(10 * time.Millisecond)
		}
	}

	log.Info(0, 0, fmt.Sprintf("maxUpload: %v", maxWrite))
	log.Info(0, 0, fmt.Sprintf("Upload: %dMB/s", maxWrite/1024/1024/3))

	// 下载测速
	maxRead := 0
	readLength, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(0, 0, "Start")
	maxRead += readLength
	startTime := time.Now()
	endTime := time.Now()

	for {
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		readLength, err = conn.Read(buf)
		if err != nil {
			log.Error(0, 0, err)
			break
		}
		_ = conn.SetReadDeadline(time.Time{})
		maxRead += readLength
		endTime = time.Now()
	}

	v := endTime.Sub(startTime) / time.Second
	log.Info(0, 0, fmt.Sprintf("maxDownload: %v", maxRead))
	log.Info(0, 0, fmt.Sprintf("v: %v", int(v)))
	log.Info(0, 0, fmt.Sprintf("Download: %dMB/s", maxRead/1024/1024/int(v)))
}
