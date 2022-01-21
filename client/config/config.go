package config

import "flag"

var (
	ServerIp   = ""
	ServerPort = 0
)

func init() {
	flag.StringVar(&ServerIp, "sip", "0.0.0.0", "server ip")
	flag.IntVar(&ServerPort, "sport", 13520, "server port")
}
