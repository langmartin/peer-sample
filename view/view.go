package view

import "math/rand"

const (
	C = 6
	H = 3
)

type Message struct {
	Addr     string
	Age      int
	InDegree int
}

type Buffer []Message

type View struct {
	Self     string
	InDegree int
	Peer     Buffer
}

// Permute shuffles the peer view window
func (v *View) Permute() {
	rand.Shuffle(len(v.Peer), func(i, j int) {
		v.Peer[i], v.Peer[j] = v.Peer[j], v.Peer[i]
	})
}

func (v *View) rmMax() Message {
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

// AgeOut moves the oldest H to the end of the window. Use inDegree as an age factor
func (v *View) AgeOut() {
	b := make(Buffer, H)
	for i := 0; i < H; i++ {
		b[i] = v.rmMax()
	}
	v.Peer = append(v.Peer, b...)
}

func (v *View) rmDups() {
	seen := make(map[string]Message)
	rm := make([]int, 0)

	for i, m := range v.Peer {
		if ok, n := seen[m.Addr]; ok {
			if m.Equal(n) {
				rm = append(rm, i)
			}
		} else {
			seen[m.Addr] = m
		}
	}
}

func (v *View) Send(peer string) Buffer {
	b := Buffer{Message{v.Self, 0, v.InDegree}}
	v.Permute()
	v.AgeOut()
	for i := 0; i < C/2-1; i++ {
		b = append(b, v.Peer[i])
	}
	return b
}

func (v *View) Recv(buf Buffer) {

}
