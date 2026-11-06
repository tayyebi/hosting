package db

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

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

// Accounts created through tenant self-registration were historically stored
// with bcrypt hashes, so verification must keep accepting them — otherwise
// those users can neither sign in nor change their password.
func TestVerifyPasswordAcceptsLegacyBcrypt(t *testing.T) {
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte("BcryptPass1"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	encoded := string(bcryptHash)
	if !strings.HasPrefix(encoded, "$2") {
		t.Fatalf("expected bcrypt hash prefix, got %q", encoded)
	}
	if !VerifyPassword(encoded, "BcryptPass1") {
		t.Fatalf("expected VerifyPassword to accept a matching bcrypt hash")
	}
	if VerifyPassword(encoded, "WrongPassword1") {
		t.Fatalf("expected VerifyPassword to reject a wrong password for a bcrypt hash")
	}
}

func TestVerifyPasswordRejectsMalformedHashes(t *testing.T) {
	for _, encoded := range []string{
		"",
		"not-a-hash",
		"pbkdf2_sha256$120000$deadbeef",       // truncated
		"pbkdf2_sha256$0$aa$bb",               // iteration count too low
		"pbkdf2_sha256$120000$zz$bb",          // invalid hex salt
		"pbkdf2_sha256$120000$aa$zz",          // invalid hex digest
		"$2a$10$invalid",                      // malformed bcrypt
		"sha256$120000$aa$bb",                 // unknown algorithm
	} {
		if VerifyPassword(encoded, "Sup3rSecret!") {
			t.Fatalf("expected VerifyPassword(%q) to fail", encoded)
		}
	}
}
