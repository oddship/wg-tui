package cache

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadState(t *testing.T) {
	dir := t.TempDir()
	want := State{
		Transfer: TransferState{
			Tool:       "rsync",
			Direction:  "upload",
			RsyncFlags: "-avz --delete",
			ScpFlags:   "-C -p",
			LocalPath:  "./dist",
			RemotePath: "/srv/app",
		},
	}

	if err := SaveState(dir, want); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	got, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if got != want {
		t.Fatalf("unexpected state:\nwant: %#v\n got: %#v", want, got)
	}
}

func TestLoadStateMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	if _, err := LoadState(dir); err == nil {
		t.Fatal("expected missing state file to fail")
	}
}
