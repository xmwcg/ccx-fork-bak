//go:build windows

package backend

import (
	"os/exec"
	"syscall"
)

func applyPlatformAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func terminateProcess(cmd *exec.Cmd) error {
	return cmd.Process.Signal(syscall.SIGTERM)
}
