package view

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

const (
	rounds    = 3000
	peers     = 300
	failure   = 0.00
	slow      = 0.00
	mortality = 0.00
)

type nodes map[string]*View

func testPush(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push()

	if rand.Float32() > failure {
		// might be dead & therefore missing
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)

			// if v.Addr == "n0" && p.Addr != "n0" {
			// 	pretty.Log("push to", b, "ended at", peer.Peer)
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
	// rand.Seed(42)

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
			testPush(p, nodes)
		}
	}

	// report OutDegree
	size := make(map[int]int)
	for _, p := range nodes {
		s := len(p.Peer)
		size[s] = size[s] + 1
	}
	ks := make([]int, 0)
	for k, _ := range size {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	for _, k := range ks {
		fmt.Printf("%02d => %d\n", k, size[k])
	}

	// report InDegree
	// indeg := make(map[int]int)
	// for _, p := range nodes {
	// 	s := len(p.Peer)
	// 	size[s] = size[s] + 1
	// }
}
