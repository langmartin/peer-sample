package view

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	rounds    = 45
	peers     = 10250
	failure   = 0.00
	replyFail = 0.00
	slow      = 0.00 // not implemented
	mortality = 0.00 // per test
)

type nodes map[string]*View
type morgue = map[string]int

// testPush implements the push only algorithm
func testPush(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push(p.Addr)
	if rand.Float32() > failure {
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)
		}
	}
	v.increaseAge()
}

// testPushPull implements the push + pull algorithm
func testPushPull(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push(p.Addr)
	if rand.Float32() > failure {
		// missing in the case the node died
		if peer, ok := ns[p.Addr]; ok {
			peer.Receive(b)
			r := peer.Push(v.Addr)
			if rand.Float32() > replyFail {
				v.Select(r)
			}
			peer.increaseAge()
		}
	}
	v.increaseAge()
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

func testRun(rounds int, peerI int, peerJ int, nodes nodes, morgue morgue) {
	c := Config{
		Size:        24,
		Heal:        8,
		Swap:        8,
		InDegreeTTL: 3,
		InDegreeAge: 1,
		CryptoRand:  false,
	}

	// init
	for i := peerI; i < peerJ; i++ {
		addr := fmt.Sprintf("n%d", i)
		boot := fmt.Sprintf("n%d", c.rint(Max(i, 1)))
		node := NewView(addr, boot)
		node.Size = c.Size
		node.Heal = c.Heal
		node.Swap = c.Swap
		node.CryptoRand = c.CryptoRand
		node.InDegreeAge = c.InDegreeAge
		node.InDegreeTTL = c.InDegreeTTL
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
}

func TestSimulation(t *testing.T) {
	nodes := make(nodes)
	morgue := make(morgue)
	// rand.Seed(4)

	// batch all the way
	// batch := 120
	// peers := 10200 // divisble by batch
	// for i := 0; i < peers; i += batch {
	// 	testRun(8, i, i+batch, nodes, morgue)
	// }

	// bootstrap
	// one := 120
	// two := (peers / 2) + 120
	// three := two + peers/2 - 120
	// testRun(8, 0, one, nodes, morgue)
	// testRun(rounds, one, two, nodes, morgue)
	// testRun(rounds, two, three, nodes, morgue)

	// full send
	testRun(rounds, 0, peers, nodes, morgue)

	// report
	indeg := countInDegree(nodes)
	fmt.Printf("alive: %d, dead: %d\n", len(nodes), len(morgue))
	fmtHist(rptOut(nodes))

	fmtHist(rptIn(nodes, indeg))
	// diameter(nodes, indeg)

	// test tests
	assert.False(t, isPartitioned(nodes), "partitioned")
}
