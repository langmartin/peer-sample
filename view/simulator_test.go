package view

import (
	"fmt"
	"math/rand"
	"testing"
)

const (
	rounds    = 100
	peers     = 10000
	failure   = 0.10
	replyFail = 0.10
	slow      = 0.00 // not implemented
	mortality = 0.10 // per test
)

type nodes map[string]*View
type morgue = map[string]int

// testPush implements the push only algorithm
func testPush(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push()
	if rand.Float32() > failure {
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)
		}
	}
}

// testPushPull implements the push + pull algorithm
func testPushPull(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push()
	if rand.Float32() < failure {
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)
			r := peer.Push()
			if rand.Float32() < replyFail {
				v.Select(r)
			}
		}
	}
}

func testKill(v *View, ns nodes, morgue morgue, time int) bool {
	mort := mortality / rounds
	if rand.Float64() < mort {
		delete(ns, v.Addr)
		morgue[v.Addr] = time
		return true
	}
	return false
}

func TestSimulation(t *testing.T) {
	nodes := make(nodes)
	morgue := make(morgue)
	// rand.Seed(427)

	boot := []string{"n0", "n1", "n2", "n3"}

	// init
	for i := 0; i < peers; i++ {
		addr := fmt.Sprintf("n%d", i)
		node := NewView(addr, boot[i%len(boot)])
		nodes[addr] = &node
	}

	// run
	for i := 0; i < rounds; i++ {
		for _, p := range nodes {
			if testKill(p, nodes, morgue, i) {
				continue
			}
			testPushPull(p, nodes)
		}
	}

	// report
	fmt.Printf("alive: %d, dead: %d\n", len(nodes), len(morgue))
	fmtHist(rptOut(nodes))
	fmtHist(rptIn(nodes))
}
