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

type EventListner struct {
	channel chan *WatcherEvent
	events  []*WatcherEvent
}

func NewEventListner() (el *EventListner) {
	el = new(EventListner)
	el.channel = make(chan *WatcherEvent)
	el.events = make([]*WatcherEvent, 0)

	go func() {
		for event := range el.channel {
			if event != nil {
				log.Debug(event)
				el.events = append(el.events, event)
			}
		}
	}()

	return
}

func (el *EventListner) Close() {
	close(el.channel)
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

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
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
	el.Close()

	if len(el.events) != 1 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 1)
	}

	if el.events[0].eventType != AddProbeDirectory {
		t.Fatalf("Invalid event type %d, expected %d", el.events[0].eventType, AddProbeDirectory)
	}

	if el.events[0].path != path {
		t.Fatalf("Invalid event path %s, expected %s", el.events[0].path, path)
	}

	if el.events[0].isNew != false {
		t.Fatalf("Invalid event path %t, expected %t", el.events[0].isNew, false)
	}
}

func TestAddProbeDirectory(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
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
	el.Close()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(el.events) != 1 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 1)
	}

	if el.events[0].eventType != AddProbeDirectory {
		t.Fatalf("Invalid event type %d, expected %d", el.events[0].eventType, AddProbeDirectory)
	}

	if el.events[0].path != path {
		t.Fatalf("Invalid event path %s, expected %s", el.events[0].path, path)
	}

	if el.events[0].isNew != true {
		t.Fatalf("Invalid event path %t, expected %t", el.events[0].isNew, true)
	}
}

func TestAddProbeDirectoryWithInvalidName(t *testing.T) {
	if err := setupWatcherTest(); err != nil {
		t.Fatal("Unable to setup test : %s", err)
	}

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
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
	el.Close()

	if len(w.directories) != 0 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 0)
	}

	if len(el.events) != 0 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 0)
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

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
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
	el.Close()

	if len(w.directories) != 0 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if len(el.events) != 2 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 2)
	}

	if el.events[1].eventType != RemoveProbeDirectory {
		t.Fatalf("Invalid event type %d, expected %d", el.events[1].eventType, RemoveProbeDirectory)
	}

	if el.events[1].path != path {
		t.Fatalf("Invalid event path %s, expected %s", el.events[1].path, path)
	}
}

func TestNewProbeDirectoryWatcherWithProbe(t *testing.T) {
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

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)
	el.Close()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(el.events) != 2 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 2)
	}

	if el.events[0].eventType != AddProbe {
		t.Fatalf("Invalid event type %d, expected %d", el.events[0].eventType, AddProbe)
	}

	if el.events[0].path != dummyProbePath {
		t.Fatalf("Invalid event path %s, expected %s", el.events[0].path, dummyProbePath)
	}

	if el.events[0].isNew != false {
		t.Fatalf("Invalid event path %t, expected %t", el.events[0].isNew, false)
	}

	if el.events[1].eventType != AddProbeDirectory {
		t.Fatalf("Invalid event type %d, expected %d", el.events[1].eventType, AddProbeDirectory)
	}

	if el.events[1].path != path {
		t.Fatalf("Invalid event path %s, expected %s", el.events[1].path, path)
	}

	if el.events[1].isNew != false {
		t.Fatalf("Invalid event path %t, expected %t", el.events[1].isNew, false)
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

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	// Create fake probe (prevent probe runner execution)
	dummyProbePath := path + "/dummy.pl"
	file, err := os.Create(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to create fake probe  %s : %s", dummyProbePath, err)
	}
	file.Close()

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)
	el.Close()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(el.events) != 2 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 2)
	}

	if el.events[1].eventType != AddProbe {
		t.Fatalf("Invalid event type %d, expected %d", el.events[1].eventType, AddProbe)
	}

	if el.events[1].path != dummyProbePath {
		t.Fatalf("Invalid event path %s, expected %s", el.events[1].path, dummyProbePath)
	}

	if el.events[1].isNew != true {
		t.Fatalf("Invalid event path %t, expected %t", el.events[1].isNew, true)
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

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

	// Remove dummy probe
	err = os.Remove(dummyProbePath)
	if err != nil {
		t.Fatalf("Unable to remove dummy probe %s : %s", dummyProbePath, err)
	}

	// Wait for fsnotify event to be triggered and processed
	time.Sleep(time.Duration(100) * time.Millisecond)
	el.Close()

	if len(w.directories) != 1 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if len(el.events) != 3 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 3)
	}

	if el.events[2].eventType != RemoveProbe {
		t.Fatalf("Invalid event type %d, expected %d", el.events[2].eventType, RemoveProbe)
	}

	if el.events[2].path != dummyProbePath {
		t.Fatalf("Invalid event path %s, expected %s", el.events[2].path, dummyProbePath)
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

	el := NewEventListner()
	w, err := NewProbeDirectoryWatcher(tmpProbeDirectory, el.channel)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Shutdown()

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
	el.Close()

	if len(w.directories) != 2 {
		t.Fatalf("Invalid probe directory count : %d, expected %d", len(w.directories), 1)
	}

	if w.directories[path].path != path {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path].path, path)
	}

	if w.directories[path2].path != path2 {
		t.Fatalf("Invalid directory path %s, expected %s", w.directories[path2].path, path2)
	}

	if len(el.events) != 5 {
		t.Fatalf("Invalid watcher event count : %d, expected %d", len(el.events), 2)
	}

	// Event order is not garenteed
	var remove bool
	var add bool
	for _, event := range el.events {
		if event.eventType == RemoveProbe && event.path == dummyProbePath {
			remove = true
		}
		if event.eventType == AddProbe && event.path == dummyProbePath2 {
			add = true
		}
	}

	if remove == false {
		t.Fatal("Missing remove event")
	}

	if add == false {
		t.Fatal("Missing add event")
	}
}
