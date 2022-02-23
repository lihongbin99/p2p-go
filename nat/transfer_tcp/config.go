package transfer_tcp

import (
	"flag"
	"fmt"
	logger "p2p-go/common/log"
	"p2p-go/common/utils"
	"strconv"
	"strings"
	"sync"
)

var (
	log             = logger.Log{From: "TCP Transfer"}
	tunS            string
	tuns            = make(map[int32]tun)
	registerSuccess = make(map[int32]tun)
	noRegister      = make(map[int32]tun)

	tunLock = sync.Mutex{}
)

type tun struct {
	remoteAddr string
	remotePort uint16
	localAddr  string
	localPort  uint16
}

func init() {
	flag.StringVar(&tunS, "tcptun", "", "ip:port=ip:port,")
}

func Init() {
	var i int32 = 1
	nameSplit := strings.Split(tunS, ",")
	for _, name := range nameSplit {
		if name == "" {
			continue
		}
		c := strings.Split(name, "=")
		if len(c) != 2 {
			log.Fatal(fmt.Errorf("config error: %s", c))
		}
		rarp := strings.Split(c[0], ":")
		if len(rarp) != 2 {
			log.Fatal(fmt.Errorf("config error: %s", rarp))
		}
		lalp := strings.Split(c[1], ":")
		if len(lalp) != 2 {
			log.Fatal(fmt.Errorf("config error: %s", lalp))
		}
		remoteAddr := rarp[0]
		if len(remoteAddr) == 0 {
			remoteAddr = "0.0.0.0"
		}
		remotePort, err := strconv.Atoi(rarp[1])
		if err != nil {
			log.Fatal(err)
		}
		localAddr := lalp[0]
		if len(localAddr) == 0 {
			localAddr = "0.0.0.0"
		}
		localPort, err := strconv.Atoi(lalp[1])
		if err != nil {
			log.Fatal(err)
		}

		tuns[i] = tun{
			remoteAddr: remoteAddr,
			remotePort: uint16(remotePort),
			localAddr:  localAddr,
			localPort:  uint16(localPort),
		}
		i++
	}

	for k, v := range tuns {
		noRegister[k] = v
	}
}

func tunToBuf(tuns map[int32]tun) (buf []byte) {
	buf = make([]byte, 0)
	for k, t := range tuns {
		id := utils.Id2Buf(k)
		buf = append(buf, id...)

		rab, err := utils.Ip2Buf(t.remoteAddr)
		if err != nil {
			log.Fatal(err)
		}
		buf = append(buf, rab...)
		rap := utils.Port2Buf(t.remotePort)
		buf = append(buf, rap...)

		lab, err := utils.Ip2Buf(t.localAddr)
		if err != nil {
			log.Fatal(err)
		}
		buf = append(buf, lab...)
		lap := utils.Port2Buf(t.localPort)
		buf = append(buf, lap...)
	}
	return
}

func BufToTun(buf []byte) (tuns map[int32]tun) {
	tuns = make(map[int32]tun)
	for len(buf) >= 16 {
		id, _ := utils.Buf2Id(buf[0:4])
		remoteAddr, _ := utils.Buf2Ip(buf[4:8])
		remotePort, _ := utils.Buf2Port(buf[8:10])
		localAddr, _ := utils.Buf2Ip(buf[10:14])
		localPort, _ := utils.Buf2Port(buf[14:16])
		tuns[id] = tun{
			remoteAddr: remoteAddr,
			remotePort: remotePort,
			localAddr:  localAddr,
			localPort:  localPort,
		}
		buf = buf[16:]
	}
	return
}
