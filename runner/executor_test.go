package runner

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"testing"
	"fmt"
	"encoding/json"
	"io"
	"errors"
	"path/filepath"
)

const dummyProbePath = "../probes/examples/dummy.pl"
const dummyProbeTmpPath = "/tmp/wigo_probe_test/dummy.pl"
const tmpProbeConfigDir = "/tmp/wigo_probe_config_test"
const dummyProbeConfigPath = "/tmp/wigo_probe_config_test/dummy.conf"

type dummyProbeConfig struct {
	Status int 		`json:"status"`
	Message string	`json:"message"`
	Exit int		`json:"exit"`
	Sleep int		`json:"sleep"`
	Stderr string	`json:"stderr"`
}

func newDummyProbeConfig(status int) (pc *dummyProbeConfig){
	pc = new(dummyProbeConfig)
	pc.Status = status
	pc.Message = "dummy"
	pc.Exit = 0
	pc.Sleep = 0
	pc.Stderr = ""
	return pc
}

func setupDummyProbeConfig(pc *dummyProbeConfig) (err error){
	// Serialize config
	json, err := json.Marshal(pc)
	if err != nil {
		return err
	}
	// Create config file
	file, err := os.OpenFile(dummyProbeConfigPath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.WriteString(file,string(json))
	if err != nil {
		return err
	}
	// Set config root
	os.Setenv("WIGO_PROBE_CONFIG_ROOT", tmpProbeConfigDir)
	return
}

func setupProbeExecutorTest() (err error) {
	// Clean everything
	log.SetLevel(log.DebugLevel)
	if err = os.RemoveAll(tmpProbeDirectory); err != nil {
		log.Errorf("Unable to remove test probe directory %s : %s", tmpProbeDirectory, err)
		return
	}
	if err = os.MkdirAll(tmpProbeDirectory, 0755); err != nil {
		log.Errorf("Unable to create test probe directory %s : %s", tmpProbeDirectory, err)
		return
	}
	if err = os.RemoveAll(tmpProbeConfigDir); err != nil {
		log.Errorf("Unable to remove test probe directory %s : %s", tmpProbeConfigDir, err)
		return
	}
	if err = os.MkdirAll(tmpProbeConfigDir, 0755); err != nil {
		log.Errorf("Unable to create test probe directory %s : %s", tmpProbeConfigDir, err)
		return
	}
	// Copy dummy probe
	cmd := exec.Command("cp", dummyProbePath, dummyProbeTmpPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf(string(output))
		err = errors.New(fmt.Sprintf("Unable to copy dummy probe from %s to %s : %s", dummyProbePath, dummyProbeTmpPath, err))
		return
	}
	// Set probe lib root
	libRoot, err := filepath.Abs("../lib")
	if err != nil {
		log.Errorf("Unable to get lib root : %s", err)
		return
	}
	os.Setenv("WIGO_PROBE_LIB_ROOT", libRoot)
	return
}

func TestExecuteProbe(t *testing.T) {
	if err := setupProbeExecutorTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}
	pe := NewProbeExecutor(dummyProbeTmpPath,1)
	result := pe.Execute()
	if result.Status != 100 {
		t.Fatalf("Invalid status %d, expected %d",result.Status,100)
	}
	if result.Message != "dummy" {
		t.Fatalf("Invalid message %s, expected %s",result.Message,"dummy")
	}
}

func TestExecuteProbeWithConfig(t *testing.T) {
	if err := setupProbeExecutorTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}
	pc := newDummyProbeConfig(226)
	pc.Message = "test"
	if err := setupDummyProbeConfig(pc) ; err != nil {
		t.Fatalf("Unable to setup dummy probe config : %s", err)
	}
	pe := NewProbeExecutor(dummyProbeTmpPath,1)
	result := pe.Execute()
	if result.Status != 226 {
		t.Fatalf("Invalid status %d, expected %d",result.Status,226)
	}
	if result.Message != "test" {
		t.Fatalf("Invalid message %s, expected %s",result.Message,"dummy")
	}
}

func TestExecuteProbeWithTimeout(t *testing.T) {
	if err := setupProbeExecutorTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}
	pe := NewProbeExecutor(dummyProbeTmpPath,1)
	pc := newDummyProbeConfig(226)
	pc.Sleep = 2
	if err := setupDummyProbeConfig(pc) ; err != nil {
		t.Fatalf("Unable to setup dummy probe config : %s", err)
	}
	result := pe.Execute()
	if result.Status != 502 {
		t.Fatalf("Invalid status %d, expected %d",result.Status,500)
	}
	if result.ExitCode != -1 {
		t.Fatalf("Invalid exit code %d, expected %d",result.ExitCode,-1)
	}
}

func TestExecuteProbeWithExitCode(t *testing.T) {
	if err := setupProbeExecutorTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}
	pe := NewProbeExecutor(dummyProbeTmpPath,1)
	pc := newDummyProbeConfig(226)
	pc.Exit = 26
	pc.Stderr = "error"
	if err := setupDummyProbeConfig(pc) ; err != nil {
		t.Fatalf("Unable to setup dummy probe config : %s", err)
	}
	result := pe.Execute()
	if result.Status != 500 {
		t.Fatalf("Invalid status %d, expected %d",result.Status,500)
	}
	if result.ExitCode != 26 {
		t.Fatalf("Invalid exit code %d, expected %d",result.ExitCode,-1)
	}
	if result.Stderr != "error" {
		t.Fatalf("Invalid stderr output %s, expected %s",result.Stderr,"error")
	}
}

func TestExecuteProbeWithDisableCode(t *testing.T) {
	if err := setupProbeExecutorTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}
	pe := NewProbeExecutor(dummyProbeTmpPath,1)
	pc := newDummyProbeConfig(226)
	pc.Exit = 13
	if err := setupDummyProbeConfig(pc) ; err != nil {
		t.Fatalf("Unable to setup dummy probe config : %s", err)
	}
	result := pe.Execute()
	if result.ExitCode != 13 {
		t.Fatalf("Invalid exit code %d, expected %d", result.ExitCode,-1)
	}
	if pe.Enabled != false {
		t.Fatalf("ProbeExecutor should have been disabled", result.ExitCode,-1)
	}
}

func TestRun(t *testing.T) {
	if err := setupProbeExecutorTest(); err != nil {
		t.Fatalf("Unable to setup test : %s", err)
	}
	pe := NewProbeExecutor(dummyProbeTmpPath,1)
	pe.Run()
}