package common

import (
	"testing"
)

func TestHashNoSalt(t *testing.T) {
	msg := "trolling"
	salt := ""
	if msg != Hash(msg, salt) {
		t.Fatal("Messages don't match")
	}
}

func TestHashSalt(t *testing.T) {
	msg := "trolling"
	salt := "myfrontlawn"
	expected := "DVgevtc9sF5A/CTus1Eaknpy2LgKzSFbFNlsM73vgsI="

	hashed := Hash(msg, salt)
	if hashed != expected {
		t.Fatalf("Hashed messages don't match")
	}
}
