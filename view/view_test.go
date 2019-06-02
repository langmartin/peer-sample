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

func TestMessageOlder(t *testing.T) {
	// by age
	m := Message{"foo", 1, 0}
	n := Message{"foo", 0, 0}
	require.True(t, m.Older(n))

	// by indegree
	m = Message{"foo", 0, 1}
	n = Message{"foo", 0, 0}
	require.True(t, m.Older(n))

	// by indegree with positive age
	m = Message{"foo", 1, 1}
	n = Message{"foo", 1, 0}
	require.True(t, m.Older(n))
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

func Test_rmDuplicates(t *testing.T) {
	// dedup exact
	n0 := NewView("n0", "n1")
	n0.Peer = append(n0.Peer, Message{"n1", 0, 0})
	n0.rmDuplicates()
	require.Equal(t, Buffer{Message{"n1", 0, 0}}, n0.Peer)

	// dedup older
	n0 = NewView("n0", "n1")
	n0.Peer = append(n0.Peer, Message{"n1", 1, 0})
	n0.rmDuplicates()
	require.Equal(t, Buffer{Message{"n1", 0, 0}}, n0.Peer)
}
