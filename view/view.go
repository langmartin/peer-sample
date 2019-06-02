package view

import (
	"math/rand"
)

const (
	Size = 3 // C
	Heal = 1 // H
	Swap = 1 // S
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

// Older checks age, adjusted by InDegree. As age grows, indegree has a smaller impact on
// the assumption that it has become less accurate
func (m *Message) Older(n Message) bool {
	age := n.Age
	if age > 0 {
		age += (n.InDegree * 1 / age)
	} else {
		age += n.InDegree
	}

	if m.Age > age {
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

func NewView(addr string, seed string) View {
	return View{addr, 0, Buffer{
		Message{seed, 0, 0},
	}}
}

func (v *View) SelectPeer() Message {
	i := rand.Intn(len(v.Peer))
	return v.Peer[i]
}

// Permute shuffles the peer view window
func (v *View) Permute() {
	rand.Shuffle(len(v.Peer), func(i, j int) {
		v.Peer[i], v.Peer[j] = v.Peer[j], v.Peer[i]
	})
}

// rmMaxAge removes the oldest message in view, using Older
func (v *View) rmMaxAge() Message {
	max := v.Peer[0]
	idx := 0
	for i := 1; i < len(v.Peer); i++ {
		if !max.Older(v.Peer[i]) {
			max = v.Peer[i]
			idx = i
		}
	}
	// FIXME could return take a list of ids to skip and append at the end? runs Heal
	// times
	v.Peer = append(v.Peer[:idx], v.Peer[idx+1:]...)
	return max
}

// AgeOut moves the oldest Heal to the end of the window
func (v *View) AgeOut() {
	b := make(Buffer, Heal)
	for i := 0; i < Heal; i++ {
		b[i] = v.rmMaxAge()
	}
	v.Peer = append(v.Peer, b...)
}

func (v *View) increaseAge() {
	for _, p := range v.Peer {
		p.Age = p.Age + 1
	}
}

// rmDuplicates keeps only the newest message for each peer
func (v *View) rmDuplicates() {
	seen := make(map[string]Message)
	rm := make([]int, 0)

	for i, m := range v.Peer {
		if n, ok := seen[m.Addr]; ok {
			if m.Older(n) {
				rm = append(rm, i) // rm m
				seen[m.Addr] = n
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
	count := Min(Heal, len(v.Peer)-Size)
	for i := 0; i < count; i++ {
		v.rmMaxAge()
	}
}

func (v *View) rmHead() {
	count := Min(Swap, len(v.Peer)-Size)
	if count > 0 {
		v.Peer = v.Peer[:count+1]
	}
}

func (v *View) rmRand() {
	count := len(v.Peer) - Size
	seen := make(map[int]bool)
	for i := 0; i < count; i++ {
		j := rand.Int()
		if _, ok := seen[j]; ok {
			continue
		}
		v.Peer = append(v.Peer[:j], v.Peer[:j+1]...)
	}
}

// Select merges the incoming buffer
func (v *View) Select(buf Buffer) {
	v.Peer = append(v.Peer, buf...)
	v.rmDuplicates()
	v.rmOld()
	v.rmHead()
	v.rmRand()
}

// ========================================================================

func (v *View) Push() Buffer {
	b := Buffer{Message{v.Addr, 0, v.InDegree}}
	v.Permute()
	v.AgeOut()
	for i := 0; i < Size/2-1; i++ {
		b = append(b, v.Peer[i])
	}
	return b
}

// func (v *View) Active() *Buffer {
// 	p := v.SelectPeer()
// 	push := true
// 	pull := false
// 	if push {
// 		return v.Push()
// 	} else {
// 		return nil
// 	}
// 	if pull {
// 	}
// }
