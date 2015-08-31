package runner

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"testing"
	"time"
)

const tmpProbeDirectory = "/tmp/wigo_probe_test"

func setupWatcherTest() (err error) {
	//	log.SetLevel(log.DebugLevel)
	if err = os.RemoveAll(tmpProbeDirectory); err != nil {
		log.Errorf("Unable to remove test probe directory %s : %s", tmpProbeDirectory, err)
		return
	}
	if err = os.Mkdir(tmpProbeDirectory, 0755); err != nil {
		log.Errorf("Unable to create test probe directory %s : %s", tmpProbeDirectory, err)
		return
	}
	return
}

type TestEventHandler struct {
	directories []string
	probes      []string
}

func NewTestEventHandler() (eh *TestEventHandler) {
	eh = new(TestEventHandler)
	eh.directories = make([]string, 0)
	eh.probes = make([]string, 0)
	return
}

func (eh *TestEventHandler) AddDirectory(path string, isNew bool) {
	log.Debugf("Add directory %s ( isNew : %t )", path, isNew)
	eh.directories = append(eh.directories, path)
}

func (eh *TestEventHandler) RemoveDirectory(path string) {
	log.Debugf("Remove directory %s", path)
	for i := 0; i < len(eh.directories); i++ {
		if eh.directories[i] == path {
			eh.directories = append(eh.directories[:i], eh.directories[i+1:]...)
			break
		}
	}
}

func (eh *TestEventHandler) AddProbe(path string, isNew bool) {
	log.Debugf("Add probe %s ( isNew : %t )", path, isNew)
	eh.probes = append(eh.probes, path)
}

func (eh *TestEventHandler) RemoveProbe(path string) {
	log.Debugf("Remove probe %s", path)
	for i := 0; i < len(eh.probes); i++ {
		if eh.probes[i] == path {
			eh.probes = append(eh.probes[:i], eh.probes[i+1:]...)
			break
		}
	}
}

func TestNewProbeDirectoryWatcher(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(eh.directories) != 1 {
		t.Fatalf("Invalid handler probe directory count: %d, expected %d", len(eh.directories), 1)
	}

	if eh.directories[0] != path {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.directories[0], path)
	}

}

func TestAddProbeDirectory(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(eh.directories) != 1 {
		t.Fatalf("Invalid handler probe directory count: %d, expected %d", len(eh.directories), 1)
	}

	if eh.directories[0] != path {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.directories[0], path)
	}
}

func TestRemoveProbeDirectory(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	// Remove probe directory
	err = os.Remove(path)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(w.directories) != 0 {
		t.Fatalf("Invalid watcher directory count : %d, expected %d", len(w.directories), 0)
	}

	if len(eh.directories) != 0 {
		t.Fatalf("Invalid handler directory count: %d, expected %d", len(eh.directories), 0)
	}
}

func TestNewProbeDirectoryWatcherWithProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Create fake probe
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(eh.directories) != 1 {
		t.Fatalf("Invalid handler directory count: %d, expected %d", len(eh.directories), 1)
	}

	if eh.directories[0] != path {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.directories[0], path)
	}

	if len(eh.probes) != 1 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 1)
	}

	if eh.probes[0] != dummyProbePath {
		t.Fatalf("Invalid handler probe path %s, expected %s", eh.probes[0], dummyProbePath)
	}

	if _, ok := w.directories[path].probes[dummyProbePath]; !ok {
		t.Fatalf("Missing probe %s from watcher", dummyProbePath)
	}
}

func TestAddProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid watcher directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid watcher directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(eh.directories) != 1 {
		t.Fatalf("Invalid handler directory count: %d, expected %d", len(eh.directories), 1)
	}

	if eh.directories[0] != path {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.directories[0], path)
	}

	// Create fake probe
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(eh.probes) != 1 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 1)
	}

	if eh.probes[0] != dummyProbePath {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.probes[0], dummyProbePath)
	}

	if _, ok := w.directories[path].probes[dummyProbePath]; !ok {
		t.Fatalf("Missing probe %s from watcher", dummyProbePath)
	}
}

func TestRemoveProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Create fake probe
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid watcher directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid watcher directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(eh.directories) != 1 {
		t.Fatalf("Invalid handler directory count: %d, expected %d", len(eh.directories), 1)
	}

	if eh.directories[0] != path {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.directories[0], path)
	}

	if len(eh.probes) != 1 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 1)
	}

	if eh.probes[0] != dummyProbePath {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.probes[0], dummyProbePath)
	}

	// Remove dummy probe
	err = os.Remove(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to remove dummy probe %s : %s", dummyProbePath, err)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(eh.probes) != 0 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 0)
	}
}

func TestRemoveProbeDirectoryWithProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Create fake probe
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid watcher directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid watcher directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(eh.directories) != 1 {
		t.Fatalf("Invalid handler directory count: %d, expected %d", len(eh.directories), 1)
	}

	if eh.directories[0] != path {
		t.Fatalf("Invalid handler directory path %s, expected %s", eh.directories[0], path)
	}

	if len(eh.probes) != 1 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 1)
	}

	if eh.probes[0] != dummyProbePath {
		t.Fatalf("Invalid handler probe path %s, expected %s", eh.probes[0], dummyProbePath)
	}

	// Remove probe directory
	err = os.RemoveAll(path)
	if err != nil {
		t.Fatalf("Unable to remove probe directory %s : %s", path, err)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(w.directories) != 0 {
		t.Fatalf("Invalid watcher directory count : %d, expected %d", len(w.directories), 0)
	}

	if len(eh.directories) != 0 {
		t.Fatalf("Invalid handler directory count: %d, expected %d", len(eh.probes), 0)
	}

	if len(eh.probes) != 0 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 0)
	}
}

func TestMoveProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Create probe directory
	path2 := tmpProbeDirectory + "/2"
	if err := os.Mkdir(path2, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path2)
	}

	// Create fake probe
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	eh := NewTestEventHandler()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, eh)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 2 {
		t.Fatalf("Invalid watcher directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if _, ok := w.directories[path].probes[dummyProbePath]; !ok {
		t.Fatalf("Missing probe %s from watcher", dummyProbePath)
	}

	if w.directories[path2].path != path2 {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path2].path, path2)
	}

	if len(eh.directories) != 2 {
		t.Fatalf("Invalid handler directory count: %d, expected %d", len(eh.directories), 1)
	}

	if len(eh.probes) != 1 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 1)
	}

	if eh.probes[0] != dummyProbePath {
		t.Fatalf("Invalid handler probe path %s, expected %s", eh.probes[0], dummyProbePath)
	}

	// Move fake probe
	dummyProbePath2 := path2 + "/dummy.pl"
	cmd := exec.Command("mv", dummyProbePath, dummyProbePath2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf(string(output))
		t.Fatalf("Unable to move dummy probe from %s to %s : %s", dummyProbePath, dummyProbePath2, err)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(eh.probes) != 1 {
		t.Fatalf("Invalid handler probe count: %d, expected %d", len(eh.probes), 1)
	}

	if eh.probes[0] != dummyProbePath2 {
		t.Fatalf("Invalid handler probe path %s, expected %s", eh.probes[0], dummyProbePath)
	}

	if _, ok := w.directories[path2].probes[dummyProbePath2]; !ok {
		t.Fatalf("Missing probe %s from watcher", dummyProbePath2)
	}
}
