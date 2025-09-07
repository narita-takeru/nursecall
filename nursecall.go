package nursecall

import (
	"os"
	"os/exec"
	"syscall"
)

const (
	statusOK = 0
)

func Start(cmdStr string, args []string, n Notifier) error {
	if err := n.Start(cmdStr); err != nil {
		return err
	}

	cmd := exec.Command(cmdStr, args...)

	exitStatus, err := do(cmd)
	if err != nil {
		return n.Error(exitStatus)
	}

	if exitStatus == statusOK {
		return n.Done(exitStatus)
	}

	// Exceptional status
	return n.Error(exitStatus)
}

func do(cmd *exec.Cmd) (int, error) {

	var err error

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return -1, err
	}

	exitStatus := statusOK
	if err := cmd.Wait(); err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			}
		}
	}
	return exitStatus, err
}
