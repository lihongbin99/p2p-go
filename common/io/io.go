package io

import (
	"net"
	"p2p-go/common/id"
	"p2p-go/common/msg"
	"p2p-go/common/utils"
	"sync"
	"time"
)

type Message struct {
	Message msg.Message
	Err     error
}

type Base struct {
	Id        int32
	Tid       int32
	lenBuf    [4]byte
	buf       []byte
	readLock  sync.Mutex
	writeLock sync.Mutex
}

func newBase() *Base {
	return &Base{id.NewId(), 0, [4]byte{}, make([]byte, 64*1024), sync.Mutex{}, sync.Mutex{}}
}
func newBaseById(id int32) *Base {
	return &Base{id, 0, [4]byte{}, make([]byte, 64*1024), sync.Mutex{}, sync.Mutex{}}
}

type TCP struct {
	*net.TCPConn
	*Base
}

func NewTCP(conn *net.TCPConn) *TCP {
	return &TCP{conn, newBase()}
}
func NewTCPById(conn *net.TCPConn, id int32) *TCP {
	return &TCP{conn, newBaseById(id)}
}

func (c *TCP) ReadMessage() Message {
	c.readLock.Lock()
	defer c.readLock.Unlock()
	maxReadLength := 0
	for maxReadLength < 4 {
		readLength, err := c.TCPConn.Read(c.lenBuf[:4-maxReadLength])
		if err != nil {
			return Message{Err: err}
		}
		maxReadLength += readLength
	}
	messageLen32, err := utils.Buf2Id(c.lenBuf[:])
	if err != nil {
		return Message{Err: err}
	}
	messageLen := int(messageLen32)

	maxReadLength = 0
	for maxReadLength < messageLen {
		readLength, err := c.TCPConn.Read(c.buf[maxReadLength : messageLen-maxReadLength])
		if err != nil {
			return Message{Err: err}
		}
		maxReadLength += readLength
	}
	message := msg.Read(c.buf[:messageLen])
	return Message{Message: message, Err: err}
}

func (c *TCP) WriteMessage(message msg.Message) (writeLen int, err error) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	buf := message.ToByteBuf()
	_, err = c.TCPConn.Write(utils.Id2Buf(int32(len(buf))))
	if err != nil {
		return
	}
	writeLen, err = c.TCPConn.Write(buf)
	return
}

type UDP struct {
	*net.UDPConn
	*Base
	LastTransferTime time.Time
	TimeOut          bool
}

func NewUDP(conn *net.UDPConn) *UDP {
	return &UDP{conn, newBase(), time.Time{}, false}
}

func NewUDPById(conn *net.UDPConn, id int32) *UDP {
	return &UDP{conn, newBaseById(id), time.Time{}, false}
}

func (c *UDP) ReadMessageFromUDP() (message msg.Message, remoteAddr *net.UDPAddr, err error) {
	c.readLock.Lock()
	defer c.readLock.Unlock()
	message = nil
	readLength, remoteAddr, err := c.UDPConn.ReadFromUDP(c.buf)
	if err != nil {
		return
	}
	message = msg.Read(c.buf[:readLength])
	return
}

func (c *UDP) WriteMessage(message msg.Message) (writeLen int, err error) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	buf := message.ToByteBuf()
	writeLen, err = c.UDPConn.Write(buf)
	return
}

func (c *UDP) WriteMessageToUDP(message msg.Message, remoteAddr *net.UDPAddr) (writeLen int, err error) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	buf := message.ToByteBuf()
	writeLen, err = c.UDPConn.WriteToUDP(buf, remoteAddr)
	return
}
