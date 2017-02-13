package nursecall

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

const (
	statusOK = 0
)

func Start(cmdStr string, args []string, n Notifier) error {
	if err := n.Validate(); err != nil {
		return err
	}

	cmd := exec.Command(cmdStr, args...)

	if err := n.Start(); err != nil {
		return err
	}

	exitStatus, err := do(cmd)
	if err != nil {
		return n.Error()
	}

	if exitStatus == statusOK {
		return n.Done()
	}

	// Exceptional status
	return n.Error()
}

func do(cmd *exec.Cmd) (int, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return -1, err
	}
	defer stdoutPipe.Close()

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return -1, err
	}
	defer stderrPipe.Close()

	if err := cmd.Start(); err != nil {
		return -1, err
	}

	go printOutput(stdoutPipe)
	go printOutput(stderrPipe)

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

func printOutput(reader io.Reader) {
	var (
		n   int
		err error
	)
	buf := make([]byte, 1024)
	for {
		if n, err = reader.Read(buf); err != nil {
			break
		}
		fmt.Println(string(buf[:n]))
	}
}
