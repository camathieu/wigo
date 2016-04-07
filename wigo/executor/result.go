package executor

import (
	"encoding/json"
	"github.com/root-gg/wigo/wigo/utils"
	pathUtil "path"
	"time"
	"path/filepath"
)

// ProbeResult is the result from a probe execution
type ProbeResult struct {
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
func NewProbeResult(status int, exitCode int, message string, details string) (pr *ProbeResult) {
	pr = new(ProbeResult)

	// Set Path and Name
//	pr.SetName(path)

	pr.Status = status
	pr.ExitCode = exitCode
	pr.Message = message
	pr.Details = details
	pr.Timestamp = time.Now().Unix()

	pr.Level = utils.StatusCodeToString(pr.Status)

	return
}

// NewProbeResultFromJSON create a new ProbeResult from JSON
func NewProbeResultFromJSON(bytes []byte) (pr *ProbeResult, err error) {
	pr = new(ProbeResult)

	err = json.Unmarshal(bytes, pr)
	if err != nil {
		return
	}

	// Override status string
	pr.Level = utils.StatusCodeToString(pr.Status)

	return
}

// ToJSON serialize a ProbeResult to json
func (pr *ProbeResult) ToJSON() (bytes []byte, err error) {
	return json.Marshal(pr)
}

//// SetName set probe name from path and remove extension if any
//func (pr *ProbeResult) SetName(path string){
//	pr.Path = path
//	fileName := pathUtil.Base(path)
//	ext := filepath.Ext(fileName)
//	pr.Name = fileName[0:len(fileName)-len(ext)]
//}

// Clean override untrusted fields
func (pr *ProbeResult) Clean(){
//	pr.Path = ""
//	pr.Name = ""
	pr.Timestamp = time.Now().Unix()
	pr.ExitCode = 0
	pr.Stdout = ""
	pr.Stderr = ""
	pr.Level = utils.StatusCodeToString(pr.Status)
}