package view

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

const (
	rounds    = 60
	peers     = 3000
	failure   = 0.00
	replyFail = 0.00
	slow      = 0.00
	mortality = 0.00
)

type hist map[int]int

func fmtHist(h hist) {
	ks := make([]int, 0)
	for k, _ := range h {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	for _, k := range ks {
		fmt.Printf("%02d => %d\n", k, h[k])
	}
}

type nodes map[string]*View

// testPush implements the push only algorithm
func testPush(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push()
	if rand.Float32() > failure {
		// might be dead & therefore missing
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)
		}
	}
}

// testPushPull implements the push + pull algorithm
func testPushPull(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push()
	if rand.Float32() > failure {
		// might be dead & therefore missing
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)
			r := peer.Push()
			if rand.Float32() > replyFail {
				v.Select(r)
			}
		}
	}
}

func testKill(v *View, ns nodes) bool {
	if rand.Float32() < mortality {
		delete(ns, v.Addr)
		return true
	}
	return false
}

func TestSimulation(t *testing.T) {
	nodes := make(nodes)
	rand.Seed(42)

	boot := []string{"n0", "n1", "n2", "n3"}

	// init
	for i := 0; i < peers; i++ {
		addr := fmt.Sprintf("n%d", i)
		node := NewView(addr, boot[i%4])
		nodes[addr] = &node
	}

	// run
	for i := 0; i < rounds; i++ {
		for _, p := range nodes {
			if testKill(p, nodes) {
				continue
			}
			testPushPull(p, nodes)
		}
	}

	// report OutDegree
	size := make(hist)
	for _, p := range nodes {
		s := len(p.Peer)
		size[s] = size[s] + 1
	}
	fmtHist(size)
}
