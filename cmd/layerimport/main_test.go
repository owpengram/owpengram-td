package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The manifest must survive a decode/encode round trip byte for byte, so that a
// Layer import only shows the lines it actually changed.
func TestWriteManifestRoundTrip(t *testing.T) {
	src := filepath.Join("..", "..", "_schema", "layers", "manifest.json")
	original, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(original, &doc); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(out, original, doc); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("manifest round trip changed the file:\n%s", got)
	}
}

func TestGitBlobID(t *testing.T) {
	// `printf 'hello\n' | git hash-object --stdin`
	if got, want := gitBlobID([]byte("hello\n")), "ce013625030ba8dba906f756967f9e9ca394464a"; got != want {
		t.Fatalf("gitBlobID = %s, want %s", got, want)
	}
}
