package global

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/root-gg/wigo/wigo/runner"
	"time"
	"sync"
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

	lock		sync.Mutex
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

// UpdateStatus recompute wigo status based on local probes status
// TODO < 100 statuses ???
func (w *Wigo) updateStatus() (status int) {
	for _, probe := range w.Probes {
		if probe.Status > status {
			status = probe.Status
		}
	}
	w.Status = status
	return
}

func (w *Wigo) RegisterProbe(probe *Probe) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.Probes[probe.Name] = probe
}

func (w *Wigo) UnregisterProbe(probe *Probe) {
	w.lock.Lock()
	defer w.lock.Unlock()

	delete(w.Probes,probe.name)
}

func (w *Wigo) UpdateProbe(result *runner.ProbeResult) (oldResult *runner.ProbeResult) {
	w.lock.Lock()
	defer w.lock.Unlock()

	log.Debugf("Got status %d for probe %s", result.Status, result.Path)
	oldResult = w.Probes[result.Name]

	// 999 is a special status to remove a probe
	if result.Status == 999 {
		delete(w.Probes, result.Name)
	} else {
		w.Probes[result.Name] = result
	}
	w.updateStatus()
	return
}

func (w *Wigo) Deduplicate(remoteWigo *Wigo) {
	for uuid, wigo := range remoteWigo.Remotes {
		if uuid != wigo.Uuid {
			log.Warnf("Remote wigo %s with uuid %s from %s mismatch ...", wigo.Hostname, wigo.Uuid, remoteWigo.Hostname)
			delete(remoteWigo.Remotes, uuid)
		}
		if w.Uuid == wigo.Uuid {
			log.Debugf("Removing local wigo from %s remotes.", remoteWigo.Hostname)
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

func (w *Wigo) UpdateRemoteWigo(remoteWigo *Wigo) (oldWigo *Wigo, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if remoteWigo.Uuid == w.Uuid {
		return
	}
	w.Deduplicate(remoteWigo)
	remoteWigo.lastUpdate = time.Now().Unix()
	oldWigo = w.Remotes[remoteWigo.Uuid]
	w.Remotes[remoteWigo.Uuid] = remoteWigo
	return
}
