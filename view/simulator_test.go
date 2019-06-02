package view

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/kr/pretty"
)

const (
	rounds    = 4
	peers     = 3
	failure   = 0.01
	slow      = 0.05
	mortality = 0.01
)

type nodes map[string]*View

func testPush(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push()

	if rand.Float32() > failure {
		// might be dead & therefore missing
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)

			// if p.Addr == "n0" {
			// 	pretty.Log(peer.Peer)
			// }
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
	node := NewView("n0", "n0")
	nodes["n0"] = &node

	// init
	for i := 1; i < peers; i++ {
		addr := fmt.Sprintf("n%d", i)
		node = NewView(addr, "n0")
		nodes[addr] = &node
	}

	// run
	for i := 0; i < rounds; i++ {
		for _, p := range nodes {
			// if p.Addr == "n0" {
			// 	pretty.Log("ROUND", p)
			// }

			if testKill(p, nodes) {
				continue
			}
			testPush(p, nodes)
		}
	}

	pretty.Log(nodes)
}
