//go:build !darwin || !cgo || offscreen || !vellum_native_checks

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "tab-close-smoke requires a Mac vellum_native_checks build")
	os.Exit(1)
}
