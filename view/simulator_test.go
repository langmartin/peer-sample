package view

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/kr/pretty"
)

const (
	rounds    = 10
	peers     = 10
	failure   = 0.01
	slow      = 0.05
	mortality = 0.01
)

type nodes map[string]View

func testPush(v View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push()
	if rand.Float32() > failure {
		// might be dead & therefore missing
		if peer, ok := ns[p.Addr]; ok {
			peer.Select(b)
		}
	}
}

func testKill(v View, ns nodes) bool {
	if rand.Float32() < mortality {
		delete(ns, v.Addr)
		return true
	}
	return false
}

func TestSimulation(t *testing.T) {
	nodes := make(nodes)
	nodes["n0"] = NewView("n0", "n1")

	// init
	for i := 1; i < peers; i++ {
		addr := fmt.Sprintf("n%d", i)
		nodes[addr] = NewView(addr, "n0")
	}

	// run
	for i := 0; i < rounds; i++ {
		fmt.Println("ROUND", i)
		for _, p := range nodes {
			if testKill(p, nodes) {
				continue
			}
			testPush(p, nodes)
		}
	}

	pretty.Log(nodes)
}
