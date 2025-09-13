package nursecall

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"syscall"
)

const (
	statusOK = 0
	maxLines = 10
	maxChars = 300
)

type Capture struct {
	lines []string
}

func (c *Capture) Write(p []byte) (int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	for scanner.Scan() {
		line := scanner.Text()
		c.lines = append(c.lines, line)
	}
	return len(p), nil
}

func (c *Capture) Summary() string {
	if len(c.lines) == 0 {
		return ""
	}

	headCount := maxLines
	if len(c.lines) < headCount {
		headCount = len(c.lines)
	}
	tailCount := maxLines
	if len(c.lines)-headCount < tailCount {
		tailCount = len(c.lines) - headCount
	}

	var buf bytes.Buffer
	// head
	for _, l := range c.lines[:headCount] {
		buf.WriteString(l + "\n")
	}
	// tail
	if tailCount > 0 {
		buf.WriteString("...\n")
		for _, l := range c.lines[len(c.lines)-tailCount:] {
			buf.WriteString(l + "\n")
		}
	}

	// 文字数制限
	s := buf.String()
	if len(s) > maxChars {
		return s[:maxChars] + "..."
	}
	return s
}

func Start(cmdStr string, args []string, n Notifier) error {
	if err := n.Start(cmdStr); err != nil {
		return err
	}

	cmd := exec.Command(cmdStr, args...)

	stdoutCap := &Capture{}
	stderrCap := &Capture{}

	cmd.Stdout = io.MultiWriter(os.Stdout, stdoutCap)
	cmd.Stderr = io.MultiWriter(os.Stderr, stderrCap)

	exitStatus, err := do(cmd)
	if err != nil {
		return n.Error(exitStatus)
	}

	n.setStdout(stdoutCap.Summary())
	n.setStderr(stderrCap.Summary())

	if exitStatus == statusOK {
		return n.Done(exitStatus)
	}
	return n.Error(exitStatus)
}

func do(cmd *exec.Cmd) (int, error) {
	if err := cmd.Start(); err != nil {
		return -1, err
	}

	exitStatus := statusOK
	if err := cmd.Wait(); err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			}
		}
		return exitStatus, err
	}
	return exitStatus, nil
}
