package handle

import (
	"p2p-go/common/io"
	"p2p-go/common/msg"
)

const (
	_ byte = iota
	Error
	Success
)

type ClientMessageHandle interface {
	ChangeStatus(cId int32, sId int32, status byte)
	Handle(cId int32, sId int32, message *io.Message) bool
}

type ServerMessageHandle interface {
	NewTCP(tcp *io.TCP, message msg.Message) bool
	NewConnect(cId int32, sId int32, writeChan chan msg.Message)
	ConnectSuccess(cId int32, sId int32, writeChan chan msg.Message)
	CloseConnect(cId int32, sId int32)
	Handle(cId int32, sId int32, ip string, port uint16, message *io.Message) (bool, bool)
}
