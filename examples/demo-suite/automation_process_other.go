//go:build !darwin

package main

import "os/exec"

// The suite's app packaging and process-group qualification target macOS.
func detachAutomationProcess(cmd *exec.Cmd) {}
