package view

import (
	"crypto/rand"
	"math/big"
)

const (
	Size        = 20 // C
	Heal        = 3  // H
	Swap        = 8  // S
	InDegreeTTL = 5
)

func rint(n int) int {
	// crypto/rand
	bn := new(big.Int).SetInt64(int64(n))
	bi, _ := rand.Int(rand.Reader, bn)
	i := int(bi.Int64())
	return i

	// math/rand
	// return rand.Intn(n)
}

type Message struct {
	Addr      string
	Age       int
	InDegree  int
	OutDegree int
}

type Buffer []*Message

func (m *Message) Equal(n Message) bool {
	return m.Addr == n.Addr &&
		m.Age == n.Age &&
		m.InDegree == n.InDegree &&
		m.OutDegree == n.OutDegree
}

// age calculates the node age, adjusted by degree
func (m *Message) ageOutDegree(c int) int {
	// OutDegree has max Size, because of window truncation. If it's much smaller than
	// Size, we want to contribute a decaying factor that keeps the node younger
	// C/out * 1/age
	o := c / (Max(m.OutDegree, 1) * Max(m.Age, 1))
	return m.Age - o
}

func (m *Message) ageInDegree(c int) int {
	// InDegree may be very large, in which case we want to be older
	// InDegree may be small, in which case we want to be younger
	// the effect shrinks with age
	if m.InDegree > c {
		return m.Age + 1
	} else if m.InDegree < c {
		return m.Age - 1
	}
	return m.Age
}

func (m *Message) age(c int) int {
	// return m.ageOutDegree(c)
	// return m.ageInDegree(c)
	return m.Age
}

// Older compares nodes by age()
func (m *Message) Older(c int, n Message) bool {
	return m.age(c) > n.age(c)
}

// ======================================================================

type LastSeen map[string]int

// View holds my own address and InDegree estimate, and the peer window
type View struct {
	// Local constants
	Size        int
	Heal        int
	Swap        int
	InDegreeTTL int

	// Real properties
	Addr     string
	Peer     Buffer
	InDegree LastSeen
}

func NewView(addr string, seed string) View {
	return View{
		Size:        Size,
		Heal:        Heal,
		Swap:        Swap,
		InDegreeTTL: InDegreeTTL,
		Addr:        addr,
		Peer:        Buffer{{Addr: seed, Age: 0, InDegree: 0, OutDegree: 1}},
		InDegree:    make(LastSeen, 0),
	}
}

func (v *View) SelectPeer() *Message {
	i := rint(len(v.Peer))
	return v.Peer[i]
}

// Permute shuffles the peer view window
func (v *View) Permute() {
	l := len(v.Peer)
	for i := l - 1; i > 0; i-- {
		j := rint(l)
		v.Peer[i], v.Peer[j] = v.Peer[j], v.Peer[i]
	}
}

// rmMaxAge removes the oldest message in view, using Older
func (v *View) rmMaxAge() *Message {
	if len(v.Peer) < 1 {
		return nil
	}
	max := v.Peer[0]
	idx := 0
	for i := 1; i < len(v.Peer); i++ {
		if !max.Older(v.Size, *v.Peer[i]) {
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
	b := make(Buffer, 0)
	var m *Message

	for i := 0; i < v.Heal; i++ {
		m = v.rmMaxAge()
		if m != nil {
			b = append(b, m)
		}
	}
	v.Peer = append(v.Peer, b...)
}

func (v *View) increaseAge() {
	for _, p := range v.Peer {
		p.Age += 1
	}
}

// Estimate my inDegree by aging a list peers that have pushed to me
func (v *View) ageInDegree(peer Message) {
	for k, t := range v.InDegree {
		if t > v.InDegreeTTL {
			delete(v.InDegree, k)
		} else {
			v.InDegree[k] = t + 1
		}
	}
	v.InDegree[peer.Addr] = 0
}

// rmDuplicates keeps only the newest message for each peer
func (v *View) rmDuplicates() {
	seen := make(map[string]*Message)
	out := make(Buffer, 0)
	for _, m := range v.Peer {
		if n, ok := seen[m.Addr]; ok {
			if m.Equal(*n) || m.Older(v.Size, *n) {
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
		j := rint(count)
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
	b := Buffer{{v.Addr, 0, len(v.InDegree), len(v.Peer)}}
	v.Permute()
	v.AgeOut()
	count := Min(v.Size/2-1, len(v.Peer))
	for i := 0; i < count; i++ {
		b = append(b, v.Peer[i])
	}
	v.increaseAge()
	return b
}

func (v *View) Receive(buf Buffer) {
	v.Select(buf)
	v.ageInDegree(*buf[0])
}
