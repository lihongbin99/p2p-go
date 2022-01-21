package main

import (
	"fmt"
	"time"
)

func main() {

	startDelayTime := time.Now()

	time.Sleep(time.Second)

	endDelayTime := time.Now()

	fmt.Printf("%dms", endDelayTime.Sub(startDelayTime)/time.Millisecond)
}
