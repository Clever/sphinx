package common

import (
	"crypto/hmac"
	"crypto/sha256"
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
	hashed := Hash(msg, salt)
	if msg == hashed {
		t.Fatal("Message matched hashed version")
	}

	hash := hmac.New(sha256.New, []byte(salt))
	hash.Write([]byte(msg))

	if !hmac.Equal([]byte(hashed), hash.Sum(nil)) {
		t.Fatal("Hashed messages don't match")
	}
}
