package converter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactInspect(t *testing.T) {
	raw := json.RawMessage(`{"Id":"abc","Config":{"Image":"nginx","Env":["DB_PASSWORD=hunter2","PATH=/usr/bin","BARE"]}}`)
	out, err := RedactInspect(raw)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "hunter2") || strings.Contains(string(out), "/usr/bin") {
		t.Fatalf("environment value survived redaction: %s", out)
	}
	var doc struct {
		ID     string `json:"Id"`
		Config struct {
			Image string   `json:"Image"`
			Env   []string `json:"Env"`
		}
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ID != "abc" || doc.Config.Image != "nginx" {
		t.Fatalf("unrelated fields changed: %s", out)
	}
	want := []string{"DB_PASSWORD=" + redacted, "PATH=" + redacted, redacted}
	for i, w := range want {
		if doc.Config.Env[i] != w {
			t.Fatalf("Env[%d] = %q, want %q", i, doc.Config.Env[i], w)
		}
	}

	// Payloads without Config or Env must pass through untouched, not error.
	for _, in := range []string{`{"Id":"abc"}`, `{"Id":"abc","Config":{"Image":"nginx"}}`} {
		if out, err := RedactInspect(json.RawMessage(in)); err != nil || string(out) != in {
			t.Fatalf("RedactInspect(%s) = %s, %v", in, out, err)
		}
	}
}
