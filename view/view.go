package view

import (
	"math/rand"
)

const (
	C    = 20
	Heal = 3
	Swap = 3
)

type Message struct {
	Addr     string
	Age      int
	InDegree int
}

type Buffer []Message

func (m *Message) Equal(n Message) bool {
	if m.Addr == n.Addr && m.Age == n.Age && m.InDegree == n.InDegree {
		return true
	}
	return false
}

// ======================================================================

// View holds my own address and InDegree estimate, and the peer window
type View struct {
	Addr     string
	InDegree int
	Peer     Buffer
}

// Permute shuffles the peer view window
func (v *View) Permute() {
	rand.Shuffle(len(v.Peer), func(i, j int) {
		v.Peer[i], v.Peer[j] = v.Peer[j], v.Peer[i]
	})
}

// rmMaxAge removes the oldest message in view. Considers InDegree
func (v *View) rmMaxAge() Message {
	var m, max Message
	max = v.Peer[0]
	idx := 0
	for i := 1; i < len(v.Peer); i++ {
		if m.Age > max.Age && (m.InDegree-10) > max.InDegree {
			max = m
			idx = i
		}
	}
	// FIXME gotta be a better way to skip
	v.Peer = append(v.Peer[:idx+1], v.Peer[idx+1:]...)
	return max
}

// AgeOut moves the oldest Heal to the end of the window. Use inDegree as an age factor
func (v *View) AgeOut() {
	b := make(Buffer, Heal)
	for i := 0; i < Heal; i++ {
		b[i] = v.rmMaxAge()
	}
	v.Peer = append(v.Peer, b...)
}

func (v *View) Send(peer string) Buffer {
	b := Buffer{Message{v.Addr, 0, v.InDegree}}
	v.Permute()
	v.AgeOut()
	for i := 0; i < C/2-1; i++ {
		b = append(b, v.Peer[i])
	}
	return b
}

// ========================================================================

func (v *View) rmDups() {
	seen := make(map[string]Message)
	rm := make([]int, 0)

	for i, m := range v.Peer {
		if n, ok := seen[m.Addr]; ok {
			if m.Equal(n) {
				rm = append(rm, i)
			}
		} else {
			seen[m.Addr] = m
		}
	}

	for _, i := range rm {
		v.Peer = append(v.Peer[:i], v.Peer[i+1:]...)
	}

}

func (v *View) rmOld() {
	count := Min(Heal, len(v.Peer)-C)
	for i := 0; i < count; i++ {
		v.rmMaxAge()
	}
}

func (v *View) rmHead() {
	count := Max(Swap, len(v.Peer)-C)
	v.Peer = v.Peer[:count+1]
}

func (v *View) rmRand() {
	count := len(v.Peer) - C
	seen := make(map[int]bool)
	for i := 0; i < count; i++ {
		j := rand.Int()
		if _, ok := seen[j]; ok {
			continue
		}
		v.Peer = append(v.Peer[:j], v.Peer[:j+1]...)
	}
}

func (v *View) Recv(buf Buffer) {

}
