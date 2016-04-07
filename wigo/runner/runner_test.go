//package runner
//
//import (
//	log "github.com/Sirupsen/logrus"
//	"os"
//	"encoding/json"
//	"io"
//	"os/exec"
//	"fmt"
//	"path/filepath"
//	"testing"
//	"time"
//)
//
//const tmpProbeDirectory1 = tmpProbeDirectory + "/1"
//
//func setupProbeRunnerTest() (err error) {
//	// Clean everything
//	if err = os.RemoveAll(tmpProbeDirectory); err != nil {
//		log.Errorf("Unable to remove test probe directory %s : %s", tmpProbeDirectory, err)
//		return
//	}
//	if err = os.MkdirAll(tmpProbeDirectory1, 0755); err != nil {
//		log.Errorf("Unable to create test probe directory %s : %s", tmpProbeDirectory1, err)
//		return
//	}
//	if err = os.RemoveAll(tmpProbeConfigDir); err != nil {
//		log.Errorf("Unable to remove test probe directory %s : %s", tmpProbeConfigDir, err)
//		return
//	}
//	if err = os.MkdirAll(tmpProbeConfigDir, 0755); err != nil {
//		log.Errorf("Unable to create test probe directory %s : %s", tmpProbeConfigDir, err)
//		return
//	}
//
//	// Set config root
//	os.Setenv("WIGO_PROBE_CONFIG_ROOT", tmpProbeConfigDir)
//
//	// Set probe lib root
//	libRoot, err := filepath.Abs("../../lib")
//	if err != nil {
//		log.Errorf("Unable to get lib root : %s", err)
//		return
//	}
//	os.Setenv("WIGO_PROBE_LIB_ROOT", libRoot)
//	return
//}
//
//func addDummyProbe(probePath string, configPath string, pc *dummyProbeConfig) (err error) {
//	// Copy dummy probe
//	cmd := exec.Command("cp", dummyProbePath, probePath)
//	output, err := cmd.CombinedOutput()
//	if err != nil {
//		log.Errorf(string(output))
//		err = fmt.Errorf("Unable to copy dummy probe from %s to %s : %s", dummyProbePath, probePath, err)
//		return
//	}
//
//	// Serialize config
//	json, err := json.Marshal(pc)
//	if err != nil {
//		return
//	}
//	// Create config file
//	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
//	if err != nil {
//		return
//	}
//	defer file.Close()
//	_, err = io.WriteString(file, string(json))
//	if err != nil {
//		return
//	}
//	return
//}
//
//
//func TestNewProbeRunner(t *testing.T) {
//	if err := setupProbeRunnerTest(); err != nil {
//		t.Fatalf("Unable to setup test : %s", err)
//	}
//	pr, err := NewProbeRunner(tmpProbeDirectory)
//	if err != nil {
//		t.Fatalf("Unable to create new ProbeRunner", err)
//	}
//	defer pr.Shutdown()
//}
//
//func TestRunProbe(t *testing.T) {
//	if err := setupProbeRunnerTest(); err != nil {
//		t.Fatalf("Unable to setup test : %s", err)
//	}
//
//	// Add dummy probe
//	pc := newDummyProbeConfig(123)
//	tmpDummyProbePath := tmpProbeDirectory1 + "/dummy1.pl"
//	err := addDummyProbe(tmpDummyProbePath, tmpProbeConfigDir + "/dummy1.conf", pc)
//	if err != nil {
//		t.Fatalf("Unable to add dummy probe : %s", err)
//	}
//
//	// Create ProbeRunner
//	pr, err := NewProbeRunner(tmpProbeDirectory)
//	if err != nil {
//		t.Fatalf("Unable to create new ProbeRunner", err)
//	}
//	defer pr.Shutdown()
//	if _, ok := pr.executors[tmpDummyProbePath] ; !ok {
//		t.Fatalf("Missing probe executor for %s", tmpDummyProbePath)
//	}
//
//	// Wait for probe result
//	select {
//	case	<- time.After(time.Second):
//		t.Fatal("Timeout waiting for probe result")
//	case	result := <- pr.Results():
//		if result.Status != 123 {
//			t.Fatal("Invalid probe status %d expected %d", result.Status, 123)
//		}
//	}
//}
//
//func TestAddProbeRunner(t *testing.T) {
//	if err := setupProbeRunnerTest(); err != nil {
//		t.Fatalf("Unable to setup test : %s", err)
//	}
//
//	// Create ProbeRunner
//	pr, err := NewProbeRunner(tmpProbeDirectory)
//	if err != nil {
//		t.Fatalf("Unable to create new ProbeRunner", err)
//	}
//	defer pr.Shutdown()
//
//	// Add dummy probe
//	pc := newDummyProbeConfig(123)
//	tmpDummyProbePath := tmpProbeDirectory1 + "/dummy1.pl"
//	err = addDummyProbe(tmpDummyProbePath, tmpProbeConfigDir + "/dummy1.conf", pc)
//	if err != nil {
//		t.Fatalf("Unable to add dummy probe : %s", err)
//	}
//
//	// Wait for fsnotify event to be triggered and processed
//	time.Sleep(time.Duration(100) * time.Millisecond)
//
//	if _, ok := pr.executors[tmpDummyProbePath] ; !ok {
//		t.Fatalf("Missing probe executor for %s", tmpDummyProbePath)
//	}
//
//	// Wait for probe result
//	select {
//	case	<- time.After(time.Second):
//		t.Fatal("Timeout waiting for probe result")
//	case	result := <- pr.Results():
//		if result.Status != 123 {
//			t.Fatal("Invalid probe status %d expected %d", result.Status, 123)
//		}
//	}
//}
//
//func TestRemoveProbeRunner(t *testing.T) {
//	if err := setupProbeRunnerTest(); err != nil {
//		t.Fatalf("Unable to setup test : %s", err)
//	}
//
//	// Add dummy probe
//	pc := newDummyProbeConfig(123)
//	tmpDummyProbePath := tmpProbeDirectory1 + "/dummy1.pl"
//	err := addDummyProbe(tmpDummyProbePath, tmpProbeConfigDir + "/dummy1.conf", pc)
//	if err != nil {
//		t.Fatalf("Unable to add dummy probe : %s", err)
//	}
//
//	// Create ProbeRunner
//	pr, err := NewProbeRunner(tmpProbeDirectory)
//	if err != nil {
//		t.Fatalf("Unable to create new ProbeRunner", err)
//	}
//	defer pr.Shutdown()
//
//	// Remove dummy probe
//	err = os.Remove(tmpDummyProbePath)
//	if err != nil {
//		t.Fatalf("Unable to remove dummy probe %s : %s", tmpDummyProbePath, err)
//	}
//
//	// Wait for probe result
//	select {
//	case	<- time.After(time.Second):
//		t.Fatal("Timeout waiting for probe result")
//	case	result := <- pr.Results():
//		if result.Status != 999 {
//			t.Fatal("Invalid probe status %d expected %d", result.Status, 999)
//		}
//	}
//
//	if _, ok := pr.executors[tmpDummyProbePath] ; ok {
//		t.Fatalf("Probe executor still present for %s", tmpDummyProbePath)
//	}
//}
//
//func TestRemoveProbeDirectoryRunner(t *testing.T) {
//	if err := setupProbeRunnerTest(); err != nil {
//		t.Fatalf("Unable to setup test : %s", err)
//	}
//
//	// Add dummy probe
//	pc := newDummyProbeConfig(123)
//	tmpDummyProbePath := tmpProbeDirectory1 + "/dummy1.pl"
//	err := addDummyProbe(tmpDummyProbePath, tmpProbeConfigDir + "/dummy1.conf", pc)
//	if err != nil {
//		t.Fatalf("Unable to add dummy probe : %s", err)
//	}
//
//	// Create ProbeRunner
//	pr, err := NewProbeRunner(tmpProbeDirectory)
//	if err != nil {
//		t.Fatalf("Unable to create new ProbeRunner", err)
//	}
//	defer pr.Shutdown()
//
//	// Remove dummy probe directory
//	err = os.RemoveAll(tmpProbeDirectory1)
//	if err != nil {
//		t.Fatalf("Unable to remove dummy probe %s : %s", tmpDummyProbePath, err)
//	}
//
//	done := make(chan struct {})
//	go func(){
//		for result := range pr.Results() {
//			log.Infof("%v",result)
//			if result.Status == 999 {
//				done <- struct {}{}
//			}
//		}
//	}()
//
//	// Wait for probe result
//	select {
//	case	<- time.After(time.Second):
//		t.Fatal("Timeout waiting for probe result")
//	case	<- done:
//		break
//	}
//
//	// Wait for fsnotify event to be triggered and processed
//	time.Sleep(time.Duration(100) * time.Millisecond)
//
//	if _, ok := pr.executors[tmpDummyProbePath] ; ok {
//		t.Fatalf("Probe executor still present for %s", tmpDummyProbePath)
//	}
//}