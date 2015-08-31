package runner

import (
	"encoding/json"
	"github.com/root-gg/wigo/wigo/utils"
	pathUtil "path"
	"time"
)

// ProbeResult is the result from a probe execution
type ProbeResult struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`

	Metrics interface{} `json:"metrics,omitempty"`
	Details interface{} `json:"details,omitempty"`

	Status   int    `json:"status"`
	Level    string `json:"level"`
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
}

// NewProbeResult create a new handcrafted ProbeResult
func NewProbeResult(path string, status int, exitCode int, message string, details string) (probeResult *ProbeResult) {
	probeResult = new(ProbeResult)

	probeResult.Path = path
	probeResult.Name = pathUtil.Base(path)

	probeResult.Status = status
	probeResult.ExitCode = exitCode
	probeResult.Message = message
	probeResult.Details = details
	probeResult.Timestamp = time.Now().Unix()

	probeResult.Level = utils.StatusCodeToString(probeResult.Status)

	return
}

// NewProbeResultFromJSON create a new ProbeResult from a probe output
func NewProbeResultFromJSON(path string, bytes []byte) (probeResult *ProbeResult, err error) {
	probeResult = new(ProbeResult)

	err = json.Unmarshal(bytes, probeResult)
	if err != nil {
		return
	}

	// Override untrusted fields
	probeResult.Path = path
	probeResult.Name = pathUtil.Base(path)
	probeResult.Timestamp = time.Now().Unix()
	probeResult.ExitCode = 0
	probeResult.Stdout = ""
	probeResult.Stderr = ""

	probeResult.Level = utils.StatusCodeToString(probeResult.Status)

	return
}

// ToJSON serialize a ProbeResult to json
func (pr *ProbeResult) ToJSON() (bytes []byte, err error) {
	return json.Marshal(pr)
}
