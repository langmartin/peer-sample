package view

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageEqual(t *testing.T) {
	m := Message{"foo", 1, 4}
	n := Message{"foo", 1, 4}
	if !m.Equal(n) {
		t.Fail()
	}
}

func TestSelect(t *testing.T) {
	n0 := NewView("n0", "n0")
	n1 := NewView("n1", "n1")

	require.Equal(t, Buffer{Message{"n1", 0, 0}}, n1.Peer)

	b := n0.Push()
	n1.Select(b)

	require.Equal(t, Buffer{
		Message{"n1", 0, 0},
		Message{"n0", 0, 0},
	},
		n1.Peer)
}
