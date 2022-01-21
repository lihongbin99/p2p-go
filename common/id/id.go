package id

import "sync"

var (
	lock sync.Mutex
	id   int32 = 0
)

func NewId() int32 {
	lock.Lock()
	defer lock.Unlock()
	id++
	return id
}
