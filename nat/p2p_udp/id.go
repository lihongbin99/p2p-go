package p2p_udp

import (
	logger "p2p-go/common/log"
	"sync"
)

var (
	log          = logger.Log{From: "UDP P2P"}
	id     int32 = 0
	idLock       = sync.Mutex{}
)

func getNewCid() (newId int32) {
	idLock.Lock()
	defer idLock.Unlock()
	id++
	newId = id
	return
}
