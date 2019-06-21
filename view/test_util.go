package view

import (
	"fmt"
	"math"
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
	mean := total / count

	var std float64
	for v, c := range h {
		std += math.Pow(float64(v-mean), 2.0) * float64(c)
	}
	std = math.Sqrt(std / float64(total))

	return avgs{mean, std}
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

func countInDegree(nodes nodes) map[string]int {
	in := make(map[string]int)
	for _, p := range nodes {
		for _, m := range p.Peer {
			in[m.Addr] = in[m.Addr] + 1
		}
	}
	return in
}

func rptIn(nodes nodes, indeg map[string]int) hist {
	h := make(hist)
	for _, c := range indeg {
		h[c] = h[c] + 1
	}
	h[0] = Max(len(nodes)-len(indeg), 0)

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
			fmt.Printf("partition found %d, missed %d\n",
				len(seen),
				len(nodes)-len(seen))
			return true
		}
	}

	return false
}

/*
func diameter(nodes nodes, indeg map[string]int) int {
	seen := make(map[string]bool)
	l := len(nodes)
	var lp func(string)
	lp = func(n string) {
		if _, ok := seen[n]; ok {
			return
		}
		// u := nodes[rint(l)]
		seen[u.Addr] = true
	}

	return 42
}
*/
