package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/root-gg/wigo/wigo/config"
	"net/http"
	"os"
)

func main() {
	log.Infof("Hello wigo")

	// Parse command line arguments
	var configFile = flag.String("config", "/etc/wigo/wigo.conf", "Configuration file (default: /etc/wigo/wigo.conf")
	flag.Parse()

	// Load config
	if err := config.LoadConfig(*configFile); err != nil {
		os.Exit(1)
	}

	// Debug mode
	if config.GetConfig().Global.Debug {
		log.SetLevel(log.DebugLevel)
		config.Dump()
		// Debug heap
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

}
