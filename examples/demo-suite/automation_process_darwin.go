package main

import (
	"os/exec"
	"syscall"
)

func detachAutomationProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
