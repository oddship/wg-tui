package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CopyProfile preserves the original files and publishes a complete copy only
// after its backup and restore instructions have been written. Existing profile
// directories are never overwritten. The caller may remove the new directory
// if saving the corresponding config fails; the backup must be retained.
func CopyProfile(root, configPath, source, destination string, state *State) (string, error) {
	// Restore instructions must work regardless of the directory wgt was run from.
	for _, path := range []*string{&root, &configPath, &source, &destination} {
		absolute, err := filepath.Abs(*path)
		if err != nil {
			return "", err
		}
		*path = absolute
	}
	if _, err := os.Lstat(destination); err == nil {
		return "", fmt.Errorf("profile data already exists at %s", destination)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	backupRoot := filepath.Join(root, ".backups")
	if err := os.MkdirAll(backupRoot, 0o700); err != nil {
		return "", err
	}
	backup, err := os.MkdirTemp(backupRoot, time.Now().UTC().Format("20060102T150405Z")+"-")
	if err != nil {
		return "", err
	}
	instructions := []string{"Close wgt before restoring. Copy these backup files to their original paths (replace existing files):"}
	for _, file := range []struct{ name, path string }{
		{"config.huml", configPath}, {"snapshot.json", filepath.Join(source, "snapshot.json")}, {"state.json", filepath.Join(source, "state.json")},
	} {
		data, err := os.ReadFile(file.path)
		if os.IsNotExist(err) && file.name != "config.huml" {
			continue
		}
		if err != nil {
			return backup, err
		}
		target := filepath.Join(backup, file.name)
		if err := os.WriteFile(target, data, 0o600); err != nil {
			return backup, err
		}
		instructions = append(instructions, fmt.Sprintf("Copy %q to %q.", target, file.path))
	}
	instructions = append(instructions, fmt.Sprintf("Then move %q aside and restart wgt. The original data at %q was retained.", destination, source))
	if err := os.WriteFile(filepath.Join(backup, "RESTORE.txt"), []byte(strings.Join(instructions, "\n")+"\n"), 0o600); err != nil {
		return backup, err
	}
	stage, err := os.MkdirTemp(root, ".profile-")
	if err != nil {
		return backup, err
	}
	defer os.RemoveAll(stage)
	for _, name := range []string{"snapshot.json", "state.json"} {
		data, err := os.ReadFile(filepath.Join(backup, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return backup, err
		}
		if err := os.WriteFile(filepath.Join(stage, name), data, 0o600); err != nil {
			return backup, err
		}
	}
	if state != nil {
		if err := SaveState(stage, *state); err != nil {
			return backup, err
		}
	}
	if err := os.Rename(stage, destination); err != nil {
		return backup, err
	}
	return backup, nil
}
