package main

import (
	"fmt"
	"ycs/core"
	"ycs/content"
)

func main() {
	// Initialize the system
	core.Initialize()
	
	// Simple test to ensure the package compiles
	doc := core.NewYDoc(core.YDocOptions{})
	fmt.Printf("Created YDoc with client ID: %d\n", doc.GetClientID())
}
