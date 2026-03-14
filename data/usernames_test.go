package data

import "testing"

func TestRandomUsername(t *testing.T) {
	if RandomUsername() == RandomUsername() {
		t.Error("this should happen eventually... but not often I hope")
	}
}

func TestRandomUsernames(t *testing.T) {
	u := RandomUsernames(10)
	if len(u) != 10 {
		t.Error("expected 10")
	}

	u = RandomUsernames(100)
	if len(u) != 100 {
		t.Error("expected 100")
	}
}
