package view

import (
	"testing"
)

func TestMessageEqual(t *testing.T) {
	m := Message{"foo", 1, 4}
	n := Message{"foo", 1, 4}
	if !m.Equal(n) {
		t.Fail()
	}
}
