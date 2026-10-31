package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func launchAutomation(dir, bundle string) (string, error) {
	file, err := os.CreateTemp(dir, "automation-workbench-*.log")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(filepath.Join(bundle, "Contents/MacOS/Vellum Workbench"), "-automation")
	cmd.Stdout = file
	cmd.Stderr = file
	detachAutomationProcess(cmd)
	if err = cmd.Start(); err != nil {
		return "", errors.Join(fmt.Errorf("start automation workbench: %w", err), file.Close())
	}
	go recordAutomationExit(cmd, file)
	return "Automation workbench launch requested. Its local socket capability is in " + file.Name(), nil
}

// The child owns duplicated descriptors. Keep this parent's descriptor until
// Wait so early failures are recorded while the hub remains alive.
func recordAutomationExit(cmd *exec.Cmd, file *os.File) {
	if err := cmd.Wait(); err != nil {
		if _, writeErr := fmt.Fprintf(file, "\nautomation workbench exited: %v\n", err); writeErr != nil {
			log.Printf("automation workbench exited: %v; diagnostic write: %v", err, writeErr)
		}
	}
	if err := file.Close(); err != nil {
		log.Printf("automation workbench log close: %v", err)
	}
}
