//go:build !js || !wasm

// Command wasm is built for js/wasm by the repository build pipeline.
package main

import "fmt"

func main() {
	fmt.Println("the wasm command must be built with GOOS=js GOARCH=wasm")
}
