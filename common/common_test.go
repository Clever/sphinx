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

func TestSortedKeys(t *testing.T) {
	obj := map[string]interface{}{
		"c": 3, "b": 2, "d": 4, "a": 1,
	}

	expected := [4]string{"a", "b", "c", "d"}
	for i, k := range SortedKeys(obj) {
		if k != expected[i] {
			t.Errorf("Expected sorted keys. Found: %#v", SortedKeys(obj))
		}
	}
}
