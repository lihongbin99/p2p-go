package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"p2p-go/common/config"
	"p2p-go/common/handle"
	mio "p2p-go/common/io"
	logger "p2p-go/common/log"
	"p2p-go/common/msg"
	"p2p-go/nat/p2p_tcp"
	"p2p-go/nat/p2p_udp"
	"strconv"
	"strings"
	"time"
)

var (
	handles = make([]handle.ServerMessageHandle, 0)
	log     = logger.Log{From: "main"}
)

func main() {
	flag.Parse()
	logger.Init()

	handles = append(handles, p2p_udp.NewServer())
	handles = append(handles, p2p_tcp.NewServer())

	localAddr, err := net.ResolveTCPAddr("tcp4", "0.0.0.0:13520")
	if err != nil {
		log.Fatal(fmt.Errorf("ResolveTCPAddr: %v", err))
	}

	listener, err := net.ListenTCP("tcp4", localAddr)
	if err != nil {
		log.Fatal(fmt.Errorf("ListenTCP: %v", err))
	}
	log.Info(0, 0, "start server success")

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Warn(0, 0, fmt.Sprintf("AcceptTCP: %v", err))
			continue
		}

		go func(tcp *mio.TCP) {
			log.Info(tcp.Tid, tcp.Id, "NewTCP")

			writeChan := make(chan msg.Message, 8)
			defer func() {
				for _, h := range handles {
					h.CloseConnect(tcp.Tid, tcp.Id)
				}
				defer close(writeChan)
				_ = tcp.Close()
				log.Info(tcp.Tid, tcp.Id, "CloseTCP")
			}()

			_ = tcp.SetReadDeadline(time.Now().Add(3 * time.Second))
			firstMessage := tcp.ReadMessage()
			if firstMessage.Err != nil {
				return
			}
			_ = tcp.SetReadDeadline(time.Time{})
			for _, h := range handles {
				if re := h.NewTCP(tcp, firstMessage.Message); re {
					return
				}
			}

			readChan := make(chan mio.Message, 8)
			readChan <- firstMessage
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

			for _, h := range handles {
				h.NewConnect(tcp.Tid, tcp.Id, writeChan)
			}

			pingTicker := time.NewTicker(time.Duration(config.KeepAliveInterval) * time.Second)
			defer pingTicker.Stop()
			lastPingTime := time.Now()
			lastPongTime := time.Now()

			var err error = nil
			for err == nil {
				select {
				case _ = <-pingTicker.C:
					lastPingTime = time.Now()
					log.Trace(tcp.Tid, tcp.Id, "send PingMessage")
					_, err = tcp.WriteMessage(&msg.PingMessage{Date: lastPingTime})
					go func() {
						time.Sleep(time.Duration(config.KeepAliveTimeOut) * time.Second)
						if lastPongTime.Before(lastPingTime) {
							log.Warn(tcp.Tid, tcp.Id, "ping timeout")
							_ = tcp.Close()
						}
					}()
				case message := <-writeChan:
					_, err = tcp.WriteMessage(message)
				case message := <-readChan:
					if message.Err != nil {
						err = message.Err
						break
					}
					switch m := message.Message.(type) {
					case *msg.ErrorMessage:
						log.Trace(tcp.Tid, tcp.Id, fmt.Sprintf("receiver ErrorMessage: %v", m.Buf))
					case *msg.PongMessage:
						log.Trace(tcp.Tid, tcp.Id, fmt.Sprintf("receiver PongMessage: %v", m.Date))
						lastPongTime = time.Now()
					case *msg.IdRequestMessage:
						tcp.Tid = m.Id
						if _, err = tcp.WriteMessage(&msg.IdResponseMessage{Id: tcp.Id}); err == nil {
							for _, h := range handles {
								h.ConnectSuccess(tcp.Tid, tcp.Id, writeChan)
							}
						}
					default:
						process := false
						remoteAddr := tcp.RemoteAddr().String()
						split := strings.Split(remoteAddr, ":")
						port, _ := strconv.Atoi(split[1])
						for _, h := range handles {
							if do, re := h.Handle(tcp, tcp.Tid, tcp.Id, split[0], uint16(port), &message); re { // core processor!
								err = io.EOF
								break
							} else if do {
								process = true
								break
							}
						}
						if !process {
							log.Trace(tcp.Tid, tcp.Id, fmt.Sprintf("no process message: %v", message.Message.ToByteBuf()))
						}
					}
				}
			}

			if err != io.EOF {
				log.Warn(tcp.Tid, tcp.Id, fmt.Sprintf("over: %v", err))
			}
		}(mio.NewTCP(conn))
	}
}
