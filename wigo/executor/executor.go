package executor

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/root-gg/utils"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
	"sync"
)

// ProbeExecutor manage running probes and
// getting probe results from them
type ProbeExecutor struct {
	Path    string
	Timeout int
	Enabled bool
	Results chan *ProbeResult

	lock	sync.Mutex
}

// NewProbeExecutor create a new ProbeExecutor instance
func NewProbeExecutor(path string, timeout int) (pe *ProbeExecutor) {
	pe = new(ProbeExecutor)
	pe.Path = path
	pe.Timeout = timeout
	pe.Enabled = true
	pe.Results = make(chan *ProbeResult)
	return pe
}

// Run a probe every delay in seconds and publish results
// to the resultChannel
func (pe *ProbeExecutor) Run() (err error) {
	for {
		timer := utils.NewSplitTime(pe.Path)
		timer.Start()
		result := pe.Execute()
		if pe.Enabled {
			pe.Results <- result
			if result.Status == 999 {
				pe.Shutdown()
				break
			}
		} else {
			break
		}
		timer.Stop()
		wait := pe.Timeout - int(timer.Elapsed().Seconds())
		if wait > 0 {
			time.Sleep(time.Duration(wait) * time.Second)
		}
	}
	return
}

// Execute the probe and always return a ProbeResult. If an error
// occurred the ProbeResult is handcrafted with the cause.
func (pe *ProbeExecutor) Execute() (probeResult *ProbeResult) {
	log.Debugf("Executing probe %s", pe.Path)

	// Stat prob
	fileInfo, err := os.Stat(pe.Path)
	if err != nil {
		log.Warnf("Failed to stat probe %s : %s", pe.Path, err)
		probeResult = NewProbeResult(999, -1, fmt.Sprintf("Failed to stat probe : %s", err), "")
		return
	}

	// Check if probe executable
	if m := fileInfo.Mode(); m&0111 == 0 {
		log.Warnf("Probe %s is not executable : %s", pe.Path, m.Perm().String())
		probeResult = NewProbeResult(998, -1, fmt.Sprintf("Probe is not executable : %s", m.Perm().String()), "")
		return
	}

	// Create command
	cmd := exec.Command(pe.Path)
	cmd.Dir = path.Dir(pe.Path)

	// Capture stdout
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Capture stderr
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Execute probe
	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case <-time.After(time.Duration(pe.Timeout) * time.Second):
		// Handle timeout
		log.Warnf("Probe %s timed out after %ds", pe.Path, pe.Timeout)
		cmd.Process.Kill()
		if err != nil {
			log.Warnf("Unable to kill probe %s : %s", pe.Path, err)
		} else {
			log.Warnf("Probe %s with pid %d killed", pe.Path, cmd.Process.Pid)
		}
		probeResult = NewProbeResult(997, -1, fmt.Sprintf("Probe timed out after %ds", pe.Timeout), "")
	case err = <-done:
		// Check if probe has been executed successfully
		if err == nil {
			// Get result from probe output
			probeResult, err = NewProbeResultFromJSON(stdout.Bytes())
			if err != nil {
				log.Warnf("Probe %s unable to deserialize probe result : %s", pe.Path, err)
				probeResult = NewProbeResult(996, -1, fmt.Sprintf("Unable to deserialize probe result : %s", err), "")
				probeResult.Stdout = string(stdout.Bytes())
				probeResult.Stderr = string(stderr.Bytes())
				return
			}
			probeResult.Clean()
		} else {
			// Get exit code
			exitCode := 1
			if exiterr, ok := err.(*exec.ExitError); ok {
				// The program has exited with an exit code != 0

				// This works on both Unix and Windows. Although package
				// syscall is generally platform dependent, WaitStatus is
				// defined for both Unix and Windows and in both cases has
				// an ExitStatus() method with the same signature.
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				}
			}

			log.Warnf("Probe %s exit code %d", pe.Path, exitCode)
			probeResult = NewProbeResult(500, exitCode, fmt.Sprintf("Exit code %d", exitCode), "")
			probeResult.Stdout = string(stdout.Bytes())
			probeResult.Stderr = string(stderr.Bytes())

			return
		}
	}

	return
}

// Shutdown disable the probe to prevent any new execution
func (pe *ProbeExecutor) Shutdown() (err error) {
	pe.Enabled = false
	return
}
