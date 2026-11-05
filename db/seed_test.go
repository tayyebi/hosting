package db

import "testing"

func TestPasswordHashingPBKDF2(t *testing.T) {
	hash, err := HashPassword("Sup3rSecret!")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !VerifyPassword(hash, "Sup3rSecret!") {
		t.Fatalf("expected VerifyPassword to succeed for matching password")
	}
	if VerifyPassword(hash, "WrongPassword1") {
		t.Fatalf("expected VerifyPassword to reject wrong password")
	}
}
