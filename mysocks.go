package main

import (
	"flag"
	"github.com/skyduy/mysocks/core"
	"log"
)

func main() {
	var mode, password, localPort, serverPort string
	flag.StringVar(&mode, "mode", "c", "running mode(s or c)")
	flag.StringVar(&password, "password", "", "password")
	flag.StringVar(&serverPort, "server-port", "", "ss-server port")
	flag.StringVar(&localPort, "local-port", "", "ss-local port")
	flag.Parse()

	if mode == "s" {
		// 启动 server 端并监听
		server, err := core.NewServer(password, serverPort)
		if err != nil {
			log.Fatalln(err)
		}
		err = server.Run()
		if err != nil {
			log.Fatalln(err)
		}
	} else if mode == "c" {
		// 启动 local 端并监听
		local, err := core.NewLocal(password, localPort, serverPort)
		if err != nil {
			log.Fatalln(err)
		}
		err = local.Run()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Fatalln("mode must be 's' or 'c'")
	}
}
