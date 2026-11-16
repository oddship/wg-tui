package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyProfileBacksUpOriginalsAndPreservesCurrentState(t *testing.T) {
	root := t.TempDir()
	source, dest := filepath.Join(root, "old"), filepath.Join(root, "new")
	config := filepath.Join(root, "custom.huml")
	if err := os.WriteFile(config, []byte("original config"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := SaveState(source, State{RecentTargets: []string{"saved"}}); err != nil {
		t.Fatal(err)
	}
	if err := Save(source, Snapshot{}); err != nil {
		t.Fatal(err)
	}
	current := State{RecentTargets: []string{"current"}, Transfer: TransferState{LocalPath: "/local", RemotePath: "/remote", RsyncFlags: "-av"}}
	backup, err := CopyProfile(root, config, source, dest, &current)
	if err != nil {
		t.Fatal(err)
	}
	copied, err := LoadState(dest)
	if err != nil || copied.RecentTargets[0] != "current" || copied.Transfer != current.Transfer {
		t.Fatalf("current state not copied: %#v %v", copied, err)
	}
	original, err := LoadState(source)
	if err != nil || original.RecentTargets[0] != "saved" {
		t.Fatalf("source modified: %#v %v", original, err)
	}
	saved, err := LoadState(backup)
	if err != nil || saved.RecentTargets[0] != "saved" {
		t.Fatalf("original state not backed up: %#v %v", saved, err)
	}
	b, err := os.ReadFile(filepath.Join(backup, "config.huml"))
	if err != nil || string(b) != "original config" {
		t.Fatalf("original config not backed up: %s %v", b, err)
	}
	instructions, err := os.ReadFile(filepath.Join(backup, "RESTORE.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{config, source, dest, filepath.Join(backup, "state.json")} {
		if !strings.Contains(string(instructions), path) {
			t.Fatalf("restore instructions omit %s: %s", path, instructions)
		}
	}
	for _, name := range []string{"config.huml", "snapshot.json", "state.json", "RESTORE.txt"} {
		info, err := os.Stat(filepath.Join(backup, name))
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("backup file permissions: %s %v", name, err)
		}
	}
	if _, err := CopyProfile(root, config, source, dest, nil); err == nil {
		t.Fatal("existing destination overwritten")
	}
}

func TestCopyProfileBackupFailureLeavesSourceAndDestinationAlone(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "config.huml")
	if err := os.WriteFile(config, []byte("original config"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := SaveState(root, State{RecentTargets: []string{"saved"}}); err != nil {
		t.Fatal(err)
	}
	// A file where the backup directory belongs gives a deterministic failure,
	// including when tests run as root.
	if err := os.WriteFile(filepath.Join(root, ".backups"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(root, "default")
	if _, err := CopyProfile(root, config, root, dest, nil); err == nil {
		t.Fatal("expected backup failure")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("destination created on failure: %v", err)
	}
	state, err := LoadState(root)
	if err != nil || state.RecentTargets[0] != "saved" {
		t.Fatalf("original state lost: %#v %v", state, err)
	}
}
