package auth

import (
	"testing"
)

func TestSetupAndUnlockRoundTrip(t *testing.T) {
	password := []byte("correct-horse-battery-staple")

	mk, rec, err := Setup(password)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer mk.Wipe()

	if len(mk.Key()) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(mk.Key()))
	}

	// Unlock with the same password should succeed.
	mk2, err := Unlock(password, rec)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	defer mk2.Wipe()

	// Both keys should be identical.
	if string(mk.Key()) != string(mk2.Key()) {
		t.Fatal("keys from Setup and Unlock differ")
	}
}

func TestUnlockWrongPassword(t *testing.T) {
	password := []byte("correct-password")
	wrong := []byte("wrong-password")

	_, rec, err := Setup(password)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	_, err = Unlock(wrong, rec)
	if err != ErrWrongPassword {
		t.Fatalf("expected ErrWrongPassword, got %v", err)
	}
}

func TestWipeZerosKey(t *testing.T) {
	mk, _, err := Setup([]byte("test"))
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	key := mk.Key()
	mk.Wipe()

	for i, b := range key {
		if b != 0 {
			t.Fatalf("byte %d not wiped: got %d", i, b)
		}
	}
}

func TestSetupProducesDifferentSalts(t *testing.T) {
	_, rec1, err := Setup([]byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	_, rec2, err := Setup([]byte("pw"))
	if err != nil {
		t.Fatal(err)
	}

	if string(rec1.Salt) == string(rec2.Salt) {
		t.Fatal("two Setup calls produced the same salt")
	}
}
