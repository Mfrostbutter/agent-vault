package pidfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteRead(t *testing.T) {
	// Use a temp dir to avoid touching real ~/.agent-vault.
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	// Ensure the .agent-vault directory exists.
	os.MkdirAll(filepath.Join(tmp, ".agent-vault"), 0700)

	pid := 12345
	if err := Write(pid); err != nil {
		t.Fatalf("Write(%d) error: %v", pid, err)
	}

	got, err := Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got != pid {
		t.Errorf("Read() = %d, want %d", got, pid)
	}
}

func TestRemove(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	os.MkdirAll(filepath.Join(tmp, ".agent-vault"), 0700)

	Write(99999)

	if err := Remove(); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	_, err := Read()
	if !os.IsNotExist(err) {
		t.Errorf("Read() after Remove() should return os.ErrNotExist, got: %v", err)
	}
}

func TestRemoveNonExistent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	os.MkdirAll(filepath.Join(tmp, ".agent-vault"), 0700)

	if err := Remove(); err != nil {
		t.Errorf("Remove() on non-existent file should not error, got: %v", err)
	}
}

func TestReadNotExist(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	os.MkdirAll(filepath.Join(tmp, ".agent-vault"), 0700)

	_, err := Read()
	if !os.IsNotExist(err) {
		t.Errorf("Read() should return os.ErrNotExist when no PID file, got: %v", err)
	}
}

func TestIsRunning(t *testing.T) {
	// Current process should be running.
	if !IsRunning(os.Getpid()) {
		t.Error("IsRunning(os.Getpid()) = false, want true")
	}

	// A very high PID should not exist.
	if IsRunning(4194304) {
		t.Error("IsRunning(4194304) = true, want false")
	}
}
