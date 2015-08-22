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

func TestNewProbeDirectoryWatcher(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Create probe directory with invalid name
	pathExamples := tmpProbeDirectory + "/examples"
	if err := os.Mkdir(pathExamples, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, pathExamples)
	}


	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
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

	if w.directories[path].time != 1 {
		t.Fatalf("Invalid directory time %d, expected %d", w.directories[path].time, 1)
	}
}

func TestAddProbeDirectory(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
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

	if w.directories[path].time != 1 {
		t.Fatalf("Invalid directory time %d, expected %d", w.directories[path].time, 1)
	}
}

func TestAddProbeDirectoryWithInvalidName(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	// Create probe directory with invalid name
	path := tmpProbeDirectory + "/test"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(w.directories) != 0 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 0)
	}
}

func TestRemoveProbeDirectory(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	// Remove probe directory
	err = os.Remove(path)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(w.directories) != 0 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 0)
	}
}

func TestNewProbDirectoryWatcherWithProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Create fake probe (prevent probe runner execution)
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	pd, ok := w.directories[path]
	if !ok {
		t.Fatalf("Missing probe directory %s", path)
	}

	if len(pd.probes) != 1 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd.probes), 1)
	}

	_, ok = pd.probes[dummyProbePath]
	if !ok {
		t.Fatalf("Missing probe %s", dummyProbePath)
	}
}

func TestAddProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	pd, ok := w.directories[path]
	if !ok {
		t.Fatalf("Missing probe directory %s", path)
	}

	// Create fake probe (prevent probe runner execution)
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(pd.probes) != 1 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd.probes), 1)
	}

	_, ok = pd.probes[dummyProbePath]
	if !ok {
		t.Fatalf("Missing probe %s", dummyProbePath)
	}
}

func TestRemoveProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	// Create probe directory
	path := tmpProbeDirectory + "/1"
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("Unable to create test probe directory %s : %s", err, path)
	}

	// Create fake probe (prevent probe runner execution)
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	pd, ok := w.directories[path]
	if !ok {
		t.Fatalf("Missing probe directory %s", path)
	}

	if len(pd.probes) != 1 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd.probes), 1)
	}

	_, ok = pd.probes[dummyProbePath]
	if !ok {
		t.Fatalf("Missing probe %s", dummyProbePath)
	}

	// Remove dummy probe
	err = os.Remove(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to remove dummy probe %s : %s", dummyProbePath, err)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)

	if len(pd.probes) != 0 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd.probes), 0)
	}
}

func TestMoveProbe(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
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

	// Create fake probe (prevent probe runner execution)
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	if len(w.directories) != 2 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 2)
	}

	pd, ok := w.directories[path]
	if !ok {
		t.Fatalf("Missing probe directory %s", path)
	}

	if len(pd.probes) != 1 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd.probes), 1)
	}

	_, ok = pd.probes[dummyProbePath]
	if !ok {
		t.Fatalf("Missing probe %s", dummyProbePath)
	}

	pd2, ok := w.directories[path2]
	if !ok {
		t.Fatalf("Missing probe directory %s", path2)
	}

	if len(pd2.probes) != 0 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd2.probes), 0)
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

	if len(pd.probes) != 0 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd.probes), 0)
	}

	if len(pd2.probes) != 1 {
		t.Fatalf("Invalid probe count : %d, expected %d", len(pd2.probes), 1)
	}

	_, ok = pd2.probes[dummyProbePath2]
	if !ok {
		t.Fatalf("Missing probe %s", dummyProbePath2)
	}
}
