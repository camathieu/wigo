package runner

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"time"
	"fmt"
	"syscall"
	"path"
	"bytes"
	"github.com/root-gg/utils"
)

type ProbeExecutor struct {
	Path  string
	Delay int
	Enabled bool
}

func NewProbeExecutor(path string, delay int) (pe *ProbeExecutor) {
	pe = new(ProbeExecutor)
	pe.Path = path
	pe.Delay = delay
	pe.Enabled = true
	return pe
}

func (pe *ProbeExecutor) Run(resultChannel chan *ProbeResult) (err error) {
	go func() {
		for {
			timer := utils.NewSplitTime(pe.Path)
			timer.Start()
			result := pe.Execute()
			if !pe.Enabled {
				break
			}
			resultChannel <- result
			timer.Stop()
			delay := pe.Delay - int(timer.Elapsed().Seconds())
			if (delay > 0) {
				time.Sleep(time.Duration(delay) * time.Second)
			}
		}
	}()
	return
}

func (pe *ProbeExecutor) Execute() (probeResult *ProbeResult) {
	log.Debugf("Executing probe %s", pe.Path)

	// Stat prob
	fileInfo, err := os.Stat(pe.Path)
	if err != nil {
		log.Warnf("Failed to stat probe %s : %s", pe.Path, err)
		NewProbeResult(pe.Path, 501, -1, fmt.Sprintf("Failed to stat probe : %s", err), "")
		return
	}

	// Check if probe executable
	if m := fileInfo.Mode(); m&0111 == 0 {
		log.Warnf("Probe %s is not executable : %s", pe.Path, m.Perm().String())
		NewProbeResult(pe.Path, 501, -1, fmt.Sprintf("Probe is not executable : %s",m.Perm().String()), "")
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
	case <-time.After(time.Duration(pe.Delay) * time.Second):
		log.Warnf("Probe %s timed out after %ds", pe.Path, pe.Delay)
		cmd.Process.Kill()
		if err != nil {
			log.Warnf("Unable to kill probe %s : %s", pe.Path, err)
		} else {
			log.Warnf("Probe %s with pid %d killed", pe.Path, cmd.Process.Pid)
		}
		probeResult = NewProbeResult(pe.Path, 502, -1, fmt.Sprintf("Probe timed out after %ds", pe.Delay), "")
	case err = <-done:
		// Check if probe has been executed sucessuflly
		if err == nil {
			// Get result from probe output
			probeResult, err = NewProbeResultFromJson(pe.Path, stdout.Bytes())
			if err != nil {
				log.Warnf("Probe %s unable to deserialize probe result : %s", pe.Path, err)
				probeResult = NewProbeResult(pe.Path, 503, -1, fmt.Sprintf("Unable to deserialize probe result : %s", err), "")
				probeResult.Stdout = string(stdout.Bytes())
				probeResult.Stderr = string(stderr.Bytes())
				return
			}
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

			if exitCode == 13 {
				log.Warnf("Probe %s exit code %d. Disabling probe", pe.Path, exitCode)
				probeResult = NewProbeResult(pe.Path, 500, exitCode, fmt.Sprintf("Exit code %d. Disabling probe", exitCode), "")
				pe.Enabled = false
			} else {
				log.Warnf("Probe %s exit code %d", pe.Path, exitCode)
				probeResult = NewProbeResult(pe.Path, 500, exitCode, fmt.Sprintf("Exit code %d", exitCode), "")
				probeResult.Stderr = string(stderr.Bytes())
			}

			return
		}
	}

	return
}

func (pe *ProbeExecutor) Shutdown() (err error) {
	pe.Enabled = false
	return
}