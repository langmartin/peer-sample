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
	failure   = 0.10
	replyFail = 0.00
	slow      = 0.00 // not implemented
	mortality = 0.10 // per test
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
}

// testPushPull implements the push + pull algorithm
func testPushPull(v *View, ns nodes) {
	p := v.SelectPeer()
	b := v.Push(p.Addr)
	if rand.Float32() > failure {
		if peer, ok := ns[p.Addr]; ok {
			peer.Receive(b)
			r := peer.Push(v.Addr)
			if rand.Float32() > replyFail {
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

func testRun(rounds int, peerI int, peerJ int, nodes nodes, morgue morgue) {
	// init
	for i := peerI; i < peerJ; i++ {
		addr := fmt.Sprintf("n%d", i)
		node := NewView(addr, "n0")
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

	// bootstrap then run
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

/*
alive: 9255, dead: 995
{2:509, 3:459, 4:378, 5:340, 6:223, 7:210, 8:164, 9:191, 10:300, 11:434, 12:497, 13:436, 14:414, 15:413, 16:441, 17:472, 18:511, 19:553, 20:2310, }
{0:0, 1:135, 2:206, 3:235, 4:349, 5:402, 6:439, 7:451, 8:495, 9:532, 10:533, 11:544, 12:567, 13:517, 14:508, 15:441, 16:369, 17:360, 18:312, 19:315, 20:274, 21:221, 22:211, 23:161, 24:135, 25:109, 26:115, 27:89, 28:73, 29:59, 30:66, 31:40, 32:36, 33:28, 34:16, 35:14, 36:15, 37:11, 38:3, 39:10, 40:2, 41:3, 42:2, 43:6, 44:1, 45:5, 46:1, 49:2, 54:1, 68:1, }
partition found 9020, missed 235


alive: 9284, dead: 966
{2:548, 3:406, 4:349, 5:366, 6:246, 7:202, 8:170, 9:203, 10:285, 11:461, 12:516, 13:429, 14:408, 15:412, 16:442, 17:506, 18:537, 19:534, 20:2264, }
{0:0, 1:123, 2:219, 3:224, 4:367, 5:374, 6:450, 7:484, 8:509, 9:499, 10:558, 11:540, 12:519, 13:488, 14:487, 15:434, 16:426, 17:387, 18:341, 19:335, 20:264, 21:236, 22:186, 23:167, 24:148, 25:128, 26:106, 27:79, 28:86, 29:51, 30:55, 31:38, 32:32, 33:16, 34:16, 35:16, 36:12, 37:11, 38:9, 39:3, 40:5, 41:2, 43:5, 45:2, 46:1, 47:2, 50:1, 51:1, 56:1, 62:1, }
partition found 9066, missed 218

alive: 9255, dead: 995
{2:532, 3:437, 4:370, 5:334, 6:226, 7:194, 8:172, 9:235, 10:300, 11:405, 12:509, 13:451, 14:404, 15:408, 16:406, 17:456, 18:499, 19:564, 20:2353, }
{0:0, 1:111, 2:211, 3:234, 4:366, 5:366, 6:426, 7:487, 8:501, 9:500, 10:544, 11:560, 12:513, 13:506, 14:461, 15:471, 16:422, 17:378, 18:369, 19:305, 20:250, 21:233, 22:207, 23:155, 24:137, 25:127, 26:82, 27:81, 28:80, 29:62, 30:49, 31:46, 32:42, 33:25, 34:10, 35:17, 36:18, 37:14, 38:5, 39:7, 40:5, 41:7, 42:2, 43:1, 44:1, 46:2, 47:1, 48:1, 49:1, 51:1, 62:1, }
partition found 9045, missed 210



*/
