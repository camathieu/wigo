package runner

import (
	"testing"
)

const validJsonResult = `
{
   "version" : "0.26",
   "status" : 226,
   "details" : {
      "foo" : "bar"
   },
   "metrics" : [
      {
         "Value" : 26,
         "Tags" : {
            "foo" : "bar"
         }
      }
   ],
   "message" : "dummy"
}
`

const invalidJsonResult = `
{
   this is invalid json
}
`

func TestNewResult(t *testing.T) {
	pr := NewProbeResult("path/to/dummy.pl", 226, 26, "this is a dummy probe", "dummy dummy dummy")

	if pr.Path != "path/to/dummy.pl" {
		t.Fatalf("Invalid probe path %s, expected %s", pr.Path, "path/to/dummy.pl")
	}

	if pr.Name != "dummy.pl" {
		t.Fatalf("Invalid probe name %s, expected %s", pr.Name, "dummy.pl")
	}

	if pr.Status != 226 {
		t.Fatalf("Invalid probe status %d, expected %d", pr.Status, 226)
	}

	if pr.Level != "WARN" {
		t.Fatalf("Invalid probe level %s, expected %s", pr.Level, "WARN")
	}

	if pr.ExitCode != 26 {
		t.Fatalf("Invalid probe exit code %d, expected %d", pr.ExitCode, 26)
	}

	if pr.Message != "this is a dummy probe" {
		t.Fatalf("Invalid probe message %s, expected %s", pr.Message, "this is a dummy probe")
	}
}

func TestNewResultFromJson(t *testing.T) {
	pr, err := NewProbeResultFromJson("path/to/dummy.pl", []byte(validJsonResult))
	if err != nil {
		t.Fatalf("Unable to deserialize valid json result : %s", err)
	}

	if pr.Status != 226 {
		t.Fatalf("Invalid probe status %d, expected %d", pr.Status, 226)
	}

	if pr.Level != "WARN" {
		t.Fatalf("Invalid probe level %s, expected %s", pr.Level, "WARN")
	}

	if pr.Message != "dummy" {
		t.Fatalf("Invalid probe message %s, expected %s", pr.Message, "this is a dummy probe")
	}

	if pr.Version != "0.26" {
		t.Fatalf("Invalid probe version %s, expected %s", pr.Version, "this is a dummy probe")
	}
}

func TestNewResultFromInvalidJson(t *testing.T) {
	_, err := NewProbeResultFromJson("path/to/dummy.pl", []byte(invalidJsonResult))
	if err != nil {
		return
	}
	t.Fatal("Deserialized  invalid json result without error")
}
