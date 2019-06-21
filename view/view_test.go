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

func TestMessage_ageInDegree(t *testing.T) {
	c := Config{Size: 10, InDegreeAge: 1}

	// no effect
	m := Message{"foo", 1, 10, 0}
	require.Equal(t, 1, m.ageInDegree(c))

	// too big
	m = Message{"foo", 1, 30, 0}
	require.Equal(t, 2, m.ageInDegree(c))

	// too small
	m = Message{"foo", 1, 2, 0}
	require.Equal(t, 0, m.ageInDegree(c))
}

func TestMessageOlder(t *testing.T) {
	c := Config{Size: 10}

	// by age
	// m := Message{"foo", 1, 0, 0}
	// n := Message{"foo", 0, 0, 0}
	// require.True(t, m.Older(10, n))

	// // by outdegree
	// m = Message{"foo", 0, 0, 10}
	// n = Message{"foo", 0, 0, 2}
	// require.True(t, m.Older(10, n))

	// by outdegree with positive age
	m := Message{"foo", 6, 0, 2}
	n := Message{"foo", 0, 0, 10}
	require.True(t, m.Older(c, n))
}

func testNode(count int) View {
	pv := make(Buffer, 0)
	for i := 1; i < count; i++ {
		l := fmt.Sprintf("n%d", i)
		pv = append(pv, &Message{l, 0, 0, 0})
	}
	n0 := NewView("n0", "")
	n0.Size = 10
	n0.Heal = 1
	n0.Swap = 1
	n0.InDegreeTTL = 5
	n0.InDegreeAge = 0
	n0.CryptoRand = false
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
	v := testNode(30)
	// Requires math/rand not crypto/rand
	rand.Seed(1)
	require.Equal(t, Buffer{
		{Addr: "n0", Age: 0, InDegree: 0, OutDegree: 29},
		{Addr: "n7", Age: 0, InDegree: 0, OutDegree: 0},
		{Addr: "n20", Age: 0, InDegree: 0, OutDegree: 0},
		{Addr: "n5", Age: 0, InDegree: 0, OutDegree: 0},
		{Addr: "n25", Age: 0, InDegree: 0, OutDegree: 0},
	},
		v.Push("n31"))
}

func TestSelect(t *testing.T) {
	n0 := NewView("n0", "n0")
	n1 := NewView("n1", "n1")

	require.Equal(t, Buffer{{"n1", 0, 0, 1}}, n1.Peer)

	b := n0.Push("n1")
	n1.Select(b)

	require.Equal(t, Buffer{
		{"n1", 0, 0, 1},
		{"n0", 0, 0, 1},
	},
		n1.Peer)
}

func TestAgeMax(t *testing.T) {
	n := NewView("n0", "n1")
	n.Peer = Buffer{
		{"n1", 3, 0, 1},
		{"n2", 2, 0, 1},
		{"n3", 1, 0, 1},
		{"n4", 4, 0, 1},
		{"n5", 4, 0, 1},
	}

	require.Equal(t, 4, n.MaxAge())

	n.Peer[2].Age = 9
	require.Equal(t, 2, n.MaxAge())
}

func Test_rmDuplicates(t *testing.T) {
	// dedup exact
	n0 := NewView("n0", "n1")
	n0.Peer = append(n0.Peer, &Message{"n1", 0, 0, 1})
	n0.rmDuplicates()
	require.Equal(t, Buffer{{"n1", 0, 0, 1}}, n0.Peer)

	// dedup older
	n0 = NewView("n0", "n1")
	n0.Peer = append(n0.Peer, &Message{"n1", 1, 0, 0})
	n0.rmDuplicates()
	require.Equal(t, Buffer{{"n1", 0, 0, 1}}, n0.Peer)

	// dedup several
	n0 = NewView("n0", "n1")
	n0.Peer = append(n0.Peer, Buffer{
		{"n1", 3, 0, 1},
		{"n2", 2, 0, 1},
		{"n2", 1, 0, 1},
		{"n2", 4, 0, 1},
	}...)
	n0.rmDuplicates()
	require.Equal(t, Buffer{
		{"n1", 0, 0, 1},
		{"n2", 1, 0, 1},
	}, n0.Peer)
}
