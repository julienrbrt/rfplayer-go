package main

import (
	"fmt"
	"os"
)

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
