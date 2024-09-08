package main

import (
	"fmt"
	"os"
)

func main() {
	if err := RFPlayerCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
