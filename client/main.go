package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"p2p-go/client/config"
	gConfig "p2p-go/common/config"
	"p2p-go/common/handle"
	mio "p2p-go/common/io"
	logger "p2p-go/common/log"
	"p2p-go/common/msg"
	"p2p-go/nat/p2p_tcp"
	"p2p-go/nat/p2p_udp"
	"strconv"
	"time"
)

var (
	log       = logger.Log{From: "main"}
	writeChan = make(chan msg.Message, 8)
	handles   = make([]handle.ClientMessageHandle, 0)
	cType     = ""
)

func init() {
	flag.StringVar(&cType, "type", "nil", "client type")
}

func notifyHandleChangeStatus(cId int32, sId int32, status byte) {
	for _, h := range handles {
		h.ChangeStatus(cId, sId, status)
	}
}

func main() {
	flag.Parse()
	logger.Init()
	switch cType {
	case "c":
		handles = append(handles, p2p_udp.NewClientClient(writeChan))
		handles = append(handles, p2p_tcp.NewClientClient(writeChan))
	case "s":
		handles = append(handles, p2p_udp.NewClientServer(writeChan))
		handles = append(handles, p2p_tcp.NewClientServer(writeChan))
	default:
		log.Fatal(fmt.Errorf("init: %s", cType))
	}

	interval := 1
	for {
		success := start()
		if success {
			interval = 1
		}
		time.Sleep(time.Duration(interval) * time.Second)
		interval = interval * 2
		if interval > 60 {
			interval = 60
		}
	}
}

func start() (success bool) {
	success = false
	serverAddr, err := net.ResolveTCPAddr("tcp4", config.ServerIp+":"+strconv.Itoa(config.ServerPort))
	if err != nil {
		log.Error(0, 0, fmt.Errorf("resolveTCPAddr: %v", err))
		return
	}

	conn, err := net.DialTCP("tcp4", nil, serverAddr)
	if err != nil {
		log.Error(0, 0, fmt.Errorf("dialTCP: %v", err))
		return
	}
	success = true
	tcp := mio.NewTCP(conn)
	log.Info(tcp.Id, tcp.Tid, "NewTCP")
	defer func() {
		notifyHandleChangeStatus(tcp.Id, tcp.Tid, handle.Error)
		log.Info(tcp.Id, tcp.Tid, "Close")
		_ = tcp.Close()
	}()

	_, err = tcp.WriteMessage(&msg.IdRequestMessage{Id: tcp.Id})
	if err != nil {
		log.Error(0, tcp.Tid, fmt.Errorf("sendId: %v", err))
		return
	}

	readChan := make(chan mio.Message, 8)
	go func() {
		defer close(readChan)
		for {
			message := tcp.ReadMessage()
			readChan <- message
			if message.Err != nil {
				break
			}
		}
	}()

	pingTicker := time.NewTicker(time.Duration(gConfig.KeepAliveInterval+gConfig.KeepAliveTimeOut) * time.Second)
	defer pingTicker.Stop()
	lastPongTime := time.Now()
	lastPingTime := time.Now()

	err = nil
	for err == nil {
		select {
		case _ = <-pingTicker.C:
			if lastPongTime.Before(lastPingTime) {
				err = fmt.Errorf("ping timeout")
			}
			lastPingTime = time.Now()
		case message := <-writeChan:
			_, err = tcp.WriteMessage(message)
		case message := <-readChan:
			if message.Err != nil {
				err = message.Err
				break
			}
			switch m := message.Message.(type) {
			case *msg.ErrorMessage:
				log.Trace(tcp.Id, tcp.Tid, fmt.Sprintf("receiver ErrorMessage: %v", m.Buf))
			case *msg.PingMessage:
				lastPongTime = time.Now()
				log.Trace(tcp.Id, tcp.Tid, fmt.Sprintf("receiver PingMessage: %v", m.Date))
				_, err = tcp.WriteMessage(&msg.PongMessage{Date: time.Now()})
			case *msg.IdResponseMessage:
				tcp.Tid = m.Id
				notifyHandleChangeStatus(tcp.Id, tcp.Tid, handle.Success)
			default:
				process := false
				for _, h := range handles {
					if h.Handle(tcp.Id, tcp.Tid, &message) { // core processor!
						process = true
						break
					}
				}
				if !process {
					log.Trace(tcp.Id, tcp.Tid, fmt.Sprintf("no process message: %v", message.Message.ToByteBuf()))
				}
			}
		}
	}

	if err != io.EOF {
		log.Warn(tcp.Id, tcp.Tid, fmt.Sprintf("over: %v", err))
	}
	return
}
