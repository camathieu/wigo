package global

import (
	"testing"
	"github.com/root-gg/wigo/wigo/runner"
)

const validJSONWigo = `
{
    "hostname": "localhost",
    "uuid": "713b75b7-c20e-45b5-bfaf-0728dd5f5ced",
    "version": "0.26",
    "alive": true,
    "status": 100,
    "probes": {
        "localdummy": {
            "path": "/usr/lib/wigo/probes/5/dummy.pl",
            "name": "localdummy",
            "version": "0.10",
            "message": "dummy",
            "timestamp": 1441139383,
            "metrics": [
                {
                    "Tags": {
                        "foo": "bar"
                    },
                    "Value": 26
                }
            ],
            "details": {
                "foo": "bar"
            },
            "status": 100,
            "level": "OK",
            "exitCode": 0
        }
    },
    "remotes": {
        "36e7706b-1d01-4357-8529-25a74126af8d": {
            "hostname": "remotehost",
            "uuid": "36e7706b-1d01-4357-8529-25a74126af8d",
            "version": "0.26",
            "alive": true,
            "status": 100,
            "probes": {
                "remotedummy": {
                    "path": "/usr/lib/wigo/probes/5/dummy.pl",
                    "name": "remotedummy",
                    "version": "0.10",
                    "message": "dummy",
                    "timestamp": 1441139383,
                    "metrics": [
                        {
                            "Tags": {
                                "foo": "bar"
                            },
                            "Value": 26
                        }
                    ],
                    "details": {
                        "foo": "bar"
                    },
                    "status": 100,
                    "level": "OK",
                    "exitCode": 0
                }
            }
        }
    }
}
`

const invalidJSONWigo = `
{
   this is invalid json
}
`

func TestNewWigo(t *testing.T){
	w := NewWigo()
	if w == nil {
		t.Fatal("Unable to create Wigo instance")
	}
}

func TestNewWigoFromJson(t *testing.T){
	w, err := NewWigoFromJson([]byte(validJSONWigo))
	if err != nil {
		t.Fatalf("Unable to load Wigo from json : %s",err)
	}

	// Check local wigo
	if w.Hostname != "localhost" {
		t.Fatalf("Invalid local wigo hostname %s, expected %s",w.Hostname, "localhost")
	}

	// Check local probes
	if w.Probes == nil {
		t.Fatal("Local probes is nil")
	}
	probe, ok := w.Probes["localdummy"]
	if !ok {
		t.Fatal("Missing localdummy probe")
	}
	if probe.Name != "localdummy" {
		t.Fatalf("Invalid probe name %s, expected %s",probe.Name, "localdummy")
	}

	// Check remote wigo
	if w.Remotes == nil {
		t.Fatal("Remotes is nil")
	}
	remoteWigo, ok := w.Remotes["36e7706b-1d01-4357-8529-25a74126af8d"]
	if !ok {
		t.Fatal("Missing remote wigo")
	}
	if remoteWigo.Hostname != "remotehost" {
		t.Fatalf("Invalid probe name %s, expected %s",remoteWigo.Hostname, "remotehost")
	}

	// Check remote probes
	if remoteWigo.Probes == nil {
		t.Fatal("Remote probes is nil")
	}
	remoteProbe, ok := remoteWigo.Probes["remotedummy"]
	if !ok {
		t.Fatal("Missing remotedummy probe")
	}
	if remoteProbe.Name != "remotedummy" {
		t.Fatalf("Invalid probe name %s, expected %s",remoteProbe.Name, "remotedummy")
	}
}

func TestNewWigoFromInvalidJson(t *testing.T){
	_, err := NewWigoFromJson([]byte(invalidJSONWigo))
	if err == nil {
		t.Fatal("No error while loading Wigo from invalid json")
	}
}

func TestUpdateProbe(t *testing.T){
	w := NewWigo()
	w.UpdateProbe(runner.NewProbeResult("/tmp/dummy.pl",226,0,"",""))
	pr, ok := w.Probes["dummy"]
	if !ok {
		t.Fatal("Missing dummy probe")
	}
	if pr.Status != 226 {
		t.Fatalf("Invalid probe status %d, expected %d", w.Status, 226)
	}
	if w.Status != 226 {
		t.Fatalf("Invalid wigo status %d, expected %d", w.Status, 226)
	}
}

func TestUpdateStatus(t *testing.T){
	w := NewWigo()
	w.UpdateProbe(runner.NewProbeResult("/tmp/dummy.pl",226,0,"",""))
	if w.Status != 226 {
		t.Fatalf("Invalid wigo status %d, expected %d", w.Status, 226)
	}
	w.UpdateProbe(runner.NewProbeResult("/tmp/dummy2.pl",100,0,"",""))
	if w.Status != 226 {
		t.Fatalf("Invalid wigo status %d, expected %d", w.Status, 226)
	}
	w.UpdateProbe(runner.NewProbeResult("/tmp/dummy3.pl",326,0,"",""))
	if w.Status != 326 {
		t.Fatalf("Invalid wigo status %d, expected %d", w.Status, 326)
	}
}

func TestRemoveProbe(t *testing.T){
	w := NewWigo()
	w.UpdateProbe(runner.NewProbeResult("/tmp/dummy.pl",100,0,"",""))
	_, ok := w.Probes["dummy"]
	if !ok {
		t.Fatal("Missing dummy probe")
	}
	w.UpdateProbe(runner.NewProbeResult("/tmp/dummy.pl",999,0,"",""))
	_, ok = w.Probes["dummy"]
	if ok {
		t.Fatal("Dummy probe has not been removed")
	}
	w.UpdateProbe(runner.NewProbeResult("/tmp/dummy2.pl",999,0,"",""))
	_, ok = w.Probes["dummy2"]
	if ok {
		t.Fatal("Dummy probe2 has not been removed")
	}
}