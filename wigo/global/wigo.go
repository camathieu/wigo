package global

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/root-gg/wigo/wigo/runner"
	"time"
)

type Wigo struct {
	Hostname   string                         `json:"hostname"`
	Uuid       string                         `json:"uuid"`
	Version    string                         `json:"version"`
	Alive      bool                           `json:"alive"`
	Status     int                            `json:"status"`
	Probes     map[string]*runner.ProbeResult `json:"probes"`
	Remotes    map[string]*Wigo               `json:"remotes"`
	lastUpdate int64
}

func NewWigo() (w *Wigo) {
	w = new(Wigo)
	w.Probes = make(map[string]*runner.ProbeResult)
	w.Remotes = make(map[string]*Wigo)
	w.Alive = true
	w.Status = 100
	return
}

func NewWigoFromJson(bytes []byte) (w *Wigo, err error) {
	w = NewWigo()
	err = json.Unmarshal(bytes, w)
	return
}

func (w *Wigo) updateStatus() (status int) {
	for _, probe := range w.Probes {
		if probe.Status > status {
			status = probe.Status
		}
	}
	w.Status = status
	return
}

func (w *Wigo) UpdateProbe(result *runner.ProbeResult) (oldResult *runner.ProbeResult, err error) {
	oldResult = w.Probes[result.Name]
	w.Probes[result.Name] = result
	w.updateStatus()
	return
}

func (w *Wigo) Deduplicate(remoteWigo *Wigo) {
	for uuid, wigo := range remoteWigo.Remotes {
		if uuid != wigo.Uuid {
			log.Warnf("Remote wigo %s with uuid %s from %s mismatch ...", wigo.Hostname(), wigo.Uuid(), remoteWigo.Hostname())
			delete(remoteWigo.Remotes, uuid)
		}
		if w.Uuid == wigo.Uuid {
			log.Debugf("Removing local wigo from %s remotes.", remoteWigo.Hostname())
			delete(remoteWigo.Remotes, uuid)
		}
		if _, ok := w.Remotes[uuid]; ok {
			log.Debugf("Removing duplicate wigo %s from %s.", wigo.Hostname, remoteWigo.Hostname)
			delete(remoteWigo.Remotes, uuid)
		}
		w.Deduplicate(wigo)
	}
	return
}

func (w *Wigo) UpdateRemoteWigo(wigo *Wigo) (oldWigo *Wigo, err error) {
	if wigo.Uuid == w.Uuid {
		return
	}
	wigo.lastUpdate = time.Now().Unix()
	oldWigo = w.Remotes[wigo.Uuid]
	w.Probes[wigo.Uuid] = wigo
	return
}
