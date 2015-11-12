package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/root-gg/wigo/wigo/config"
	"net/http"
	"os"
	"github.com/root-gg/wigo/wigo/runner"
	"github.com/root-gg/wigo/wigo/global"
	"github.com/root-gg/utils"
)

var wigo *global.Wigo

func main() {
	log.Info("Hello wigo")

	// Parse command line arguments
	var configFile = flag.String("config", "/etc/wigo/wigo.conf", "Configuration file (default: /etc/wigo/wigo.conf)")
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

	// Create local wigo
	wigo = global.NewWigo()
	wigo.Hostname = config.GetConfig().Global.Hostname

	// Start local probe runner
	pr, err := runner.NewProbeRunner(config.GetConfig().Global.ProbesDirectory)
	if err != nil {
		log.Warn("Unable to start local probe runner : %s", err)
		os.Exit(1)
	}

	// Handle local probe results
	go func(){
		for {
			result := <- pr.Results()
			oldResult := wigo.UpdateProbe(result)
			compareProbeResults(oldResult,result)
		}
	}()

	select{}
}

func compareProbeResults(old *runner.ProbeResult, new *runner.ProbeResult){
	if old.Status != new.Status {
//		notify.Handle(old,new)
//		log.Handle(old,new)
	}
}

func compareWigo(old *global.Wigo, new *global.Wigo){

}

func saveMetrics(pr *runner.ProbeResult){

}