//go:build ignore
// +build ignore

package main

import (
	"fmt"
)

func main() {
	home, _ := ccLensHome()
	items, status := parseHermes(home)
	toks := 0
	cost := 0.0
	for _, it := range items {
		toks += it.Tokens
		cost += it.Cost
	}
	fmt.Printf("status=%s records=%d sessions=%d toks=%d cost=%.12f\n", status.State, status.Records, len(uniqueSessionIDs(items)), toks, cost)
}
