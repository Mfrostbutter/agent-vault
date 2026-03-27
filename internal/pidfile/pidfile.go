package pidfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const fileName = "agent-vault.pid"

// Path returns the path to the PID file (~/.agent-vault/agent-vault.pid).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agent-vault", fileName), nil
}

// Write atomically writes the given PID to the PID file.
func Write(pid int) error {
	path, err := Path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(pid)+"\n"), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Read reads the PID from the PID file.
// Returns os.ErrNotExist if the file does not exist.
func Read() (int, error) {
	path, err := Path()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file content: %w", err)
	}
	return pid, nil
}

// Remove deletes the PID file. No error is returned if the file does not exist.
func Remove() error {
	path, err := Path()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// IsRunning checks whether a process with the given PID is still running.
func IsRunning(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
