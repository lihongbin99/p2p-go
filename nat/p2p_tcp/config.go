package p2p_tcp

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var (
	nameS   string
	nameMap = make(map[string]int)
	names   = make([]string, 0)
)

func init() {
	flag.StringVar(&nameS, "tcp", "", "name:port,")
}

func Init() {
	nameSplit := strings.Split(nameS, ",")
	for _, name := range nameSplit {
		np := strings.Split(name, ":")
		if len(np) != 2 {
			log.Fatal(fmt.Errorf("config error: %s", np))
		}
		if port, err := strconv.Atoi(np[1]); err != nil {
			log.Fatal(err)
		} else {
			nameMap[np[0]] = port
		}
	}

	for name := range nameMap {
		names = append(names, name)
	}
}
