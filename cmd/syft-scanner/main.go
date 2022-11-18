package main

import (
	"github.com/jedevc/buildkit-syft-scanner/internal"
)

func main() {
	// TODO: enable logging

	scanner, err := internal.NewScannerFromEnvironment()
	if err != nil {
		panic(err)
	}

	if err := scanner.Scan(); err != nil {
		panic(err)
	}
}
