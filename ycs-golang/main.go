package main

import (
	"fmt"
	"ycs/contracts"
	"ycs/core"
)

func main() {
	// Initialize the system
	core.Initialize()

	// Simple test to ensure the package compiles
	doc := core.NewYDoc(contracts.YDocOptions{})
	fmt.Printf("Created YDoc with client ID: %d\n", doc.GetClientID())
}
