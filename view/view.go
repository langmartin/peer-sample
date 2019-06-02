package view

import (
	"math/rand"
)

const (
	Size = 10 // C
	Heal = 1  // H
	Swap = 1  // S
)

type Message struct {
	Addr     string
	Age      int
	InDegree int
}

type Buffer []*Message

func (m *Message) Equal(n Message) bool {
	if m.Addr == n.Addr && m.Age == n.Age && m.InDegree == n.InDegree {
		return true
	}
	return false
}

// Older checks age, adjusted by InDegree. As age grows, indegree has a smaller impact on
// the assumption that it has become less accurate
func (m *Message) Older(n Message) bool {
	// calculate m's age
	am := m.Age
	am += m.InDegree * (1 / Max(1, am))

	// and n's age
	an := n.Age
	an += n.InDegree * (1 / Max(1, an))

	if am > an {
		return true
	}
	return false
}

// ======================================================================

// View holds my own address and InDegree estimate, and the peer window
type View struct {
	Size     int
	Heal     int
	Swap     int
	Addr     string
	InDegree int
	Peer     Buffer
}

func NewView(addr string, seed string) View {
	return View{Size, Heal, Swap, addr, 0, Buffer{
		&Message{seed, 0, 0},
	}}
}

func (v *View) SelectPeer() *Message {
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
func (v *View) rmMaxAge() *Message {
	max := v.Peer[0]
	idx := 0
	for i := 1; i < len(v.Peer); i++ {
		if !max.Older(*v.Peer[i]) {
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
	b := make(Buffer, v.Heal)
	for i := 0; i < v.Heal; i++ {
		b[i] = v.rmMaxAge()
	}
	v.Peer = append(v.Peer, b...)
}

func (v *View) increaseAge() {
	for _, p := range v.Peer {
		p.Age += 1
	}
}

// rmDuplicates keeps only the newest message for each peer
func (v *View) rmDuplicates() {
	seen := make(map[string]*Message)
	out := make(Buffer, 0)
	for _, m := range v.Peer {
		if n, ok := seen[m.Addr]; ok {
			if m.Equal(*n) || m.Older(*n) {
				seen[m.Addr] = n
			} else {
				out = append(out, m)
			}
		} else {
			seen[m.Addr] = m
			out = append(out, m)
		}
	}
	v.Peer = out
}

func (v *View) rmOld() {
	count := Min(v.Heal, len(v.Peer)-v.Size)
	for i := 0; i < count; i++ {
		v.rmMaxAge()
	}
}

func (v *View) rmHead() {
	count := Min(v.Swap, len(v.Peer)-v.Size)
	if count > 0 {
		v.Peer = v.Peer[:count+1]
	}
}

func (v *View) rmRand() {
	count := len(v.Peer) - v.Size
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
	v.increaseAge()
}

// ========================================================================

func (v *View) Push() Buffer {
	b := Buffer{&Message{v.Addr, 0, v.InDegree}}
	v.Permute()
	v.AgeOut()
	count := Min(v.Size/2-1, len(v.Peer))
	for i := 0; i < count; i++ {
		b = append(b, v.Peer[i])
	}
	v.increaseAge()
	return b
}
