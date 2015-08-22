package runner

import (
	log "github.com/Sirupsen/logrus"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"os"
	pathUtil "path"
	"strconv"
)

// ProbeDirectoryWatcher watchs for probe directories.
// Each probe directory in this directory is expected to
// be named by a number specifying the delay in seconds between
// two executions of the probes that it contains. Other files or
// directories will not be added ( eg : the examples folder ).
type ProbeDirectoryWatcher struct {
	path        string
	directories map[string]*ProbeDirectory
	watcher     *fsnotify.Watcher
	stop        chan struct{}
}

// NewProbeDirectoryWatcher creates a new ProbeDirectoryWatcher instance
func NewProbeDirectoryWatcher(path string) (w *ProbeDirectoryWatcher, err error) {
	log.Debug("New probe directory watcher : " + path)

	w = new(ProbeDirectoryWatcher)
	w.directories = make(map[string]*ProbeDirectory)
	w.stop = make(chan struct{})
	w.path = path

	// Check if the probe directory exist
	src, err := os.Stat(w.path)
	if err != nil {
		log.Errorf("Probe directory %s does not exists : %s", path, err)
		return
	}

	// Check if the probe directory is indeed a directory
	if !src.IsDir() {
		log.Errorf("Probe directory %s is not a directory : %s", path, err)
		return
	}

	// Start watcher first to be sure we don't miss any event
	if err = w.watch(); err != nil {
		return
	}

	// Read directory
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Errorf("Unable to list probe directories %s : %s", path, err)
		return
	}

	// Keep only directories
	for _, f := range files {
		if f.IsDir() {
			w.addDirectory(w.path + "/" + f.Name())
		}
	}

	return
}

// Watch starts watching the probe directory
func (w *ProbeDirectoryWatcher) watch() (err error) {
	// Create fsnotify watcher
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("Unable to create fsnotify watcher for %s : %s", w.path, err)
		return
	}

	err = w.watcher.Watch(w.path)
	if err != nil {
		log.Errorf("Unable to create fsnotify watcher on %s : %s", w.path, err)
		return
	}

	// Watch for changes forever
	go func() {
	loop:
		for {
			select {
			case <-w.stop:
				// Shutdown gracefully
				w.watcher.Close()
				break loop
			case ev := <-w.watcher.Event:
				if ev.IsCreate() {
					fileInfo, err := os.Stat(ev.Name)
					if err != nil {
						log.Errorf("Error stating %s : %s", ev.Name, err)
						continue
					}
					if fileInfo.IsDir() {
						w.addDirectory(ev.Name)
					}
				} else if ev.IsDelete() {
					w.removeDirectory(ev.Name)
				} else if ev.IsRename() {
					w.removeDirectory(ev.Name)
				}
			case err := <-w.watcher.Error:
				log.Warn("%s fsnotify watcher error : %s", w.path, err)
			}

		}
	}()

	return
}

// Shutdown gracefully and recursively stops all directory watchers
// and probe runners
func (w *ProbeDirectoryWatcher) Shutdown() (err error) {
	log.Debug("Shutdown probe directory watcher : " + w.path)
	w.stop <- struct{}{}
	for _, pd := range w.directories {
		pd.shutdown()
	}
	return
}

// Add a new probe directory to watch
func (w *ProbeDirectoryWatcher) addDirectory(path string) (err error) {
	// Check if directory exists already
	if _, ok := w.directories[path]; ok {
		log.Warn("Probe directory %s has already been added. Discarding", path)
		return
	}

	// Extract execution delay from directory name
	dirname := pathUtil.Base(path)
	delay, err := strconv.Atoi(dirname)
	if err != nil {
		if dirname != "examples" {
			log.Warnf("Probe directory %s is not numeric. Discarding.", dirname)
		}
		return
	}

	// Create ProbeDirectory
	pd, err := newProbeDirectory(path, delay)
	if err != nil {
		return
	}
	w.directories[path] = pd

	return
}

// RemoveDirectory removes a probe directory from the watcher.
// Usually when a probe directory is removed from the file system.
func (w *ProbeDirectoryWatcher) removeDirectory(path string) {
	log.Debug("Remove probe directory : " + path)
	if path == w.path {
		log.Warn("Probe directory %s has been removed. Shutting down probe watcher", w.path)
		w.Shutdown()
		return
	}

	if _, ok := w.directories[path]; !ok {
		log.Warnf("Probe directory %s is not present. Discarding", path)
		return
	}

	w.directories[path].shutdown()
	delete(w.directories, path)
	return
}

// ProbeDirectory watch over a probe directory.
// It is expected to contain runnable wigo probes.
type ProbeDirectory struct {
	path    string
	time    int
	probes  map[string]*ProbeExecutor
	watcher *fsnotify.Watcher
	stop    chan struct{}
}

// NewProbeDirectory create a new ProbeDirectory instance
func newProbeDirectory(path string, time int) (pd *ProbeDirectory, err error) {
	log.Debug("New probe directory : " + path)

	pd = new(ProbeDirectory)
	pd.path = path
	pd.time = time
	pd.probes = make(map[string]*ProbeExecutor)
	pd.stop = make(chan struct{})

	// check if the probe directory exist
	src, err := os.Stat(pd.path)
	if err != nil {
		log.Errorf("Probe directory %s does not exists : %s", path, err)
		return
	}

	// check if the probe directory is a directory
	if !src.IsDir() {
		log.Errorf("Probe directory %s is not a directory : %s", path, err)
		return
	}

	if err = pd.watch(); err != nil {
		return
	}

	// Read directory
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Errorf("Unable to list probe directory %s : %s", path, err)
		return
	}

	// Keep only files
	for _, f := range files {
		if !f.IsDir() {
			pd.addProbe(pd.path + "/" + f.Name())
		}
	}

	return
}

// Watch starts watching the probe directory
func (pd *ProbeDirectory) watch() (err error) {
	// Create fsnotify watcher
	pd.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("Unable to create fsnotify watcher for %s : %s", pd.path, err)
		return
	}

	// Create a watcher on probe directory
	err = pd.watcher.Watch(pd.path)
	if err != nil {
		log.Errorf("Unable to create fsnotify watcher on %s : %s", pd.path, err)
		return
	}

	// Watch for changes forever
	go func() {
	loop:
		for {
			select {
			case <-pd.stop:
				// Graceful shutdown
				pd.watcher.Close()
				break loop
			case ev := <-pd.watcher.Event:
				if ev.IsCreate() {
					fileInfo, err := os.Stat(ev.Name)
					if err != nil {
						log.Errorf("Error stating %s : %s", ev.Name, err)
						continue
					}
					if !fileInfo.IsDir() {
						pd.addProbe(ev.Name)
					}
				} else if ev.IsDelete() {
					pd.removeProbe(ev.Name)
				} else if ev.IsRename() {
					pd.removeProbe(ev.Name)
				}
			case err := <-pd.watcher.Error:
				log.Warn("%s fsnotify watcher error : %s", pd.path, err)
			}

		}
	}()

	return
}

// Shutdown gracefully and recursively stops all directory watchers
// and probe runners.
func (pd *ProbeDirectory) shutdown() (err error) {
	log.Debug("Shutdown probe directory : " + pd.path)
	pd.stop <- struct{}{}
	for _, pr := range pd.probes {
		pr.Shutdown()
	}
	return
}

// AddProbe adds a new ProbeExecutor for the probe.
func (pd *ProbeDirectory) addProbe(path string) (err error) {
	log.Debug("Add probe : " + path)

	// Check if probe already exists
	if _, ok := pd.probes[path]; ok {
		log.Warnf("Probe %s already exists", path)
	}

	pr := NewProbeExecutor(path, pd.time)
	err = pr.Run()
	if err != nil {
		return
	}
	pd.probes[path] = pr
	return
}

// RemoveProbe removes a probe from the ProbeDirectory.
// Usually when a probe is removed from the file system.
func (pd *ProbeDirectory) removeProbe(path string) (err error) {
	// It's not possible to stat a removed path
	if path == pd.path {
		return
	}
	log.Debug("Remove probe : " + path)

	// Check if probe exists
	pr, ok := pd.probes[path]
	if !ok {
		log.Warnf("Probe %s is not present. Discarding", path)
		return
	}

	pr.Shutdown()
	delete(pd.probes, path)
	return
}
