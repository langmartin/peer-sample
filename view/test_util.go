package view

import (
	"fmt"
	"os"
	"sort"
)

type hist map[int]int

func fmtHist(h hist) {
	ks := make([]int, 0)
	for k, _ := range h {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	fmt.Print("{")
	for _, k := range ks {
		fmt.Printf("%d:%d, ", k, h[k])
	}
	fmt.Println("}")
}

type avgs struct {
	mean int
	std  float64
}

func avgHist(h hist) avgs {
	count, total := 0, 0
	for v, c := range h {
		count = count + c
		total = total + v*c
	}
	std := 0.1
	return avgs{total / count, std}
}

func fmtR(file string, xss [][]int) {
	f, _ := os.Create(file)
	defer f.Close()
	f.WriteString("#!/usr/bin/env Rscript\n")
	f.WriteString("pdf(file=\"" + file + ".pdf\")")
	f.WriteString("plot")
}

func rptOut(nodes nodes) hist {
	size := make(hist)
	for _, p := range nodes {
		s := len(p.Peer)
		size[s] = size[s] + 1
	}
	return size
}

func indeg(nodes nodes) map[string]int {
	in := make(map[string]int)
	for _, p := range nodes {
		for _, m := range p.Peer {
			in[m.Addr] = in[m.Addr] + 1
		}
	}
	return in
}

func rptIn(nodes nodes) hist {
	in := indeg(nodes)
	h := make(hist)
	for _, c := range in {
		h[c] = h[c] + 1
	}
	h[0] = len(nodes) - len(in)

	// // print zero nodes
	// if h[0] > 0 {
	// 	for n, _ := range nodes {
	// 		if _, ok := in[n]; !ok {
	// 			fmt.Printf("zero indegree %s\n", n)
	// 		}
	// 	}
	// }

	return h
}

// isPartitioned does a DFS of the graph of nodes to ensure that all nodes are connected
func isPartitioned(nodes nodes) bool {
	seen := make(map[string]bool)
	var lp func(string)
	lp = func(n string) {
		if _, ok := seen[n]; ok {
			return
		}

		// someone has a dead node still in view, which is fine
		if _, ok := nodes[n]; !ok {
			return
		}

		seen[n] = true
		for _, m := range nodes[n].Peer {
			lp(m.Addr)
		}
	}

	// populate seen
	lp("n0")

	for n, _ := range nodes {
		if _, ok := seen[n]; !ok {
			fmt.Printf("partition found %d\n", len(seen))
			return true
		}
	}

	return false
}