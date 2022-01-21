package p2p_tcp

import (
	logger "p2p-go/common/log"
	"sync"
)

var (
	log          = logger.Log{From: "TCP P2P"}
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
