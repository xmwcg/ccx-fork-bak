//go:build !windows

package backend

import (
	"os"
	"os/exec"
)

func applyPlatformAttrs(cmd *exec.Cmd) {}

func terminateProcess(cmd *exec.Cmd) error {
	return cmd.Process.Signal(os.Interrupt)
}
