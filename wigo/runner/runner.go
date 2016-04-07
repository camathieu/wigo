//package runner
//
//import (
//	log "github.com/Sirupsen/logrus"
//	pathUtil "path"
//	"strconv"
//	"sync"
//)
//
//type ProbeRunner struct {
//	watcher       *ProbeDirectoryWatcher
//	executors     map[string]*ProbeExecutor
//	resultChannel chan *ProbeResult
//	lock          sync.Mutex
//}
//
//func NewProbeRunner(probeDirectory string) (pr *ProbeRunner, err error){
//	pr = new(ProbeRunner)
//	pr.resultChannel = make(chan *ProbeResult)
//	pr.executors = make(map[string]*ProbeExecutor)
//	pr.watcher, err = NewProbeDirectoryWatcher(probeDirectory, pr)
//	return
//}
//
//func (pr *ProbeRunner) Results() chan *ProbeResult {
//	return pr.resultChannel
//}
//
//func (pr *ProbeRunner) AddDirectory(path string, isNew bool) {
//	log.Infof("Adding probe directory %s", path)
//}
//
//func (pr *ProbeRunner) RemoveDirectory(path string) {
//	log.Infof("Removing probe directory %s", path)
//}
//
//func (pr *ProbeRunner) AddProbe(path string, isNew bool) {
//	log.Infof("Adding probe executor for %s", path)
//
//	pr.lock.Lock()
//	defer pr.lock.Unlock()
//
//	// Verify directory name
//	dirname := pathUtil.Base(pathUtil.Dir(path))
//	timeout, err := strconv.Atoi(dirname)
//	if err != nil {
//		if dirname != "examples" {
//			log.Warnf("Probe directory %s is not numeric. Discarding.", dirname)
//		}
//		return
//	}
//
//	if _, ok := pr.executors[path]; ok {
//		log.Warn("Executor for probe %s already exists", path)
//		return
//	}
//
//
//	pe := NewProbeExecutor(path, timeout)
//	pe.Run(pr.resultChannel)
//	pr.executors[path] = pe
//}
//
//func (pr *ProbeRunner) RemoveProbe(path string) {
//	log.Infof("Removing probe executor for %s", path)
//
//	pr.lock.Lock()
//	defer pr.lock.Unlock()
//
//	if _, ok := pr.executors[path]; ok {
//		pr.executors[path].Shutdown()
//		delete(pr.executors, path)
//		pr.resultChannel <- NewProbeResult(path, 999, -1, "Probe has been removed", "")
//		return
//	}
//
//	log.Warn("Executor for probe %s does not exist", path)
//}
//
//func (pr *ProbeRunner) Shutdown() {
//	pr.watcher.Shutdown()
//	for _, pe := range pr.executors {
//		pe.Shutdown()
//	}
//	close(pr.resultChannel)
//}