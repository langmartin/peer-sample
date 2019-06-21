package view

import (
	cr "crypto/rand"
	"math/big"
	"math/rand"
)

// ======================================================================

const (
	Size        = 20 // C
	Heal        = 3  // H
	Swap        = 8  // S
	InDegreeTTL = 5
	InDegreeAge = 1
)

type Config struct {
	Size        int
	Heal        int
	Swap        int
	InDegreeTTL int
	InDegreeAge int
	CryptoRand  bool
}

func (c *Config) rint(n int) int {
	if c.CryptoRand {
		bn := new(big.Int).SetInt64(int64(n))
		bi, _ := cr.Int(cr.Reader, bn)
		i := int(bi.Int64())
		return i
	}
	// math/rand
	return rand.Intn(n)
}

// ======================================================================

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
func (m *Message) ageOutDegree(c Config) int {
	// OutDegree has max Size, because of window truncation. If it's much smaller than
	// Size, we want to contribute a decaying factor that keeps the node younger
	// C/out * 1/age
	o := c.Size / (Max(m.OutDegree, 1) * Max(m.Age, 1))
	return m.Age - o
}

func (m *Message) ageInDegree(c Config) int {
	// InDegree may be very large, in which case we want to be older
	// InDegree may be small, in which case we want to be younger
	// the effect shrinks with age
	if m.InDegree > c.Size {
		return m.Age + c.InDegreeAge
	} else if m.InDegree < c.Size {
		return m.Age - c.InDegreeAge
	}
	return m.Age
}

func (m *Message) age(c Config) int {
	// return m.ageOutDegree(c)
	// return m.ageInDegree(c)
	return m.Age
}

// Older compares nodes by age()
func (m *Message) Older(c Config, n Message) bool {
	return m.age(c) > n.age(c)
}

// ======================================================================

type LastSeen map[string]int

// View holds my own address and InDegree estimate, and the peer window
type View struct {
	Config
	Addr     string
	Peer     Buffer
	InDegree LastSeen
}

func NewView(addr string, seed string) View {
	return View{
		Config: Config{
			Size:        Size,
			Heal:        Heal,
			Swap:        Swap,
			InDegreeTTL: InDegreeTTL,
			InDegreeAge: InDegreeAge,
			CryptoRand:  true,
		},
		Addr:     addr,
		Peer:     Buffer{{Addr: seed, Age: 0, InDegree: 0, OutDegree: 1}},
		InDegree: make(LastSeen, 0),
	}
}

func (v *View) SelectPeer() *Message {
	// c := v.Config
	// i := c.rint(len(v.Peer))
	i := v.MaxAge()
	return v.Peer[i]
}

func (v *View) rmPeer(i int) {
	v.Peer = append(v.Peer[:i], v.Peer[i+1:]...)
}

// Permute shuffles the peer view window
func (v *View) Permute() {
	c := v.Config
	l := len(v.Peer)
	for i := l - 1; i > 0; i-- {
		j := c.rint(i)
		v.Peer[i], v.Peer[j] = v.Peer[j], v.Peer[i]
	}
}

// MaxAge finds the index of the oldest peer
func (v *View) MaxAge() int {
	c := v.Config
	if len(v.Peer) < 1 {
		return -1
	}
	max := v.Peer[0]
	idx := 0
	for i := 1; i < len(v.Peer); i++ {
		if !max.Older(c, *v.Peer[i]) {
			max = v.Peer[i]
			idx = i
		}
	}
	return idx
}

// rmMaxAge removes the oldest message in view, using Older
func (v *View) rmMaxAge() *Message {
	i := v.MaxAge()
	if i < 0 {
		return nil
	}
	node := v.Peer[i]
	v.rmPeer(i)
	return node
}

// AgeOut moves the oldest Heal to the end of the window
func (v *View) AgeOut() {
	b := Buffer{}
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
	c := v.Config

	// make the list of messages to keep
	keep := make(map[string]int)
	for i, m := range v.Peer {
		if j, ok := keep[m.Addr]; ok {
			// this test is backwards
			// m is the new candidate and Peer[j] is current
			if !m.Older(c, *v.Peer[j]) {
				keep[m.Addr] = i
			}
		} else {
			keep[m.Addr] = i
		}
	}

	out := Buffer{}
	for i, m := range v.Peer {
		if i == keep[m.Addr] {
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
	c := v.Config
	l := len(v.Peer)
	count := l - v.Size
	for i := 0; i < count; i++ {
		v.rmPeer(c.rint(l - i))
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

func (v *View) Push(peerAddr string) Buffer {
	b := Buffer{{v.Addr, 0, len(v.InDegree), len(v.Peer)}}
	v.Permute()
	v.AgeOut()
	count := Min(v.Size/2-1, len(v.Peer))
	for i := 0; i < count; i++ {
		// Don't send a peer its own record
		if v.Peer[i].Addr == peerAddr {
			continue
		}
		b = append(b, v.Peer[i])
	}
	return b
}

func (v *View) Receive(buf Buffer) {
	v.Select(buf)
	v.ageInDegree(*buf[0])
}
