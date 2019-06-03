package view

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageEqual(t *testing.T) {
	m := Message{"foo", 1, 4, 1}
	n := Message{"foo", 1, 4, 1}
	if !m.Equal(n) {
		t.Fail()
	}
}

func TestMessageOlder(t *testing.T) {
	// by age
	m := Message{"foo", 1, 0, 0}
	n := Message{"foo", 0, 0, 0}
	require.True(t, m.Older(n))

	// by indegree
	m = Message{"foo", 0, 1, 0}
	n = Message{"foo", 0, 0, 0}
	require.True(t, m.Older(n))

	// by indegree with positive age
	m = Message{"foo", 1, 1, 0}
	n = Message{"foo", 1, 0, 0}
	require.True(t, m.Older(n))
}

func testNode(count int) View {
	pv := make(Buffer, 0)
	for i := 0; i < count; i++ {
		l := fmt.Sprintf("n%d", i)
		pv = append(pv, &Message{l, 0, 0, 0})
	}
	n0 := NewView("n0", "")
	n0.Peer = pv
	return n0
}

func Test_increaseAge(t *testing.T) {
	n := NewView("n0", "n1")
	n.increaseAge()
	n.increaseAge()
	require.Equal(t, Buffer{{Addr: "n1", Age: 2, InDegree: 0, OutDegree: 1}}, n.Peer)
}

func TestPush(t *testing.T) {
	rand.Seed(1)
	v := testNode(30)
	require.Equal(t, Buffer{
		{Addr: "n0", Age: 0, InDegree: 0, OutDegree: 30},
		{Addr: "n9", Age: 1, InDegree: 0, OutDegree: 0},
		{Addr: "n14", Age: 1, InDegree: 0, OutDegree: 0},
		{Addr: "n0", Age: 1, InDegree: 0, OutDegree: 0},
		{Addr: "n25", Age: 1, InDegree: 0, OutDegree: 0},
	},
		v.Push())
}

func TestSelect(t *testing.T) {
	n0 := NewView("n0", "n0")
	n1 := NewView("n1", "n1")

	require.Equal(t, Buffer{{"n1", 0, 0, 0}}, n1.Peer)

	b := n0.Push()
	n1.Select(b)

	require.Equal(t, Buffer{
		{"n1", 1, 0, 0},
		{"n0", 1, 0, 0},
	},
		n1.Peer)
}

func Test_rmDuplicates(t *testing.T) {
	// dedup exact
	n0 := NewView("n0", "n1")
	n0.Peer = append(n0.Peer, &Message{"n1", 0, 0, 0})
	n0.rmDuplicates()
	require.Equal(t, Buffer{{"n1", 0, 0, 0}}, n0.Peer)

	// dedup older
	n0 = NewView("n0", "n1")
	n0.Peer = append(n0.Peer, &Message{"n1", 1, 0, 0})
	n0.rmDuplicates()
	require.Equal(t, Buffer{{"n1", 0, 0, 0}}, n0.Peer)
}
