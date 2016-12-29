package nursecall

import (
	"io"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"strconv"
	"os"
)

func Start(cmdStr string, args []string) {
	cmd := exec.Command(cmdStr, args...)

	n := notifier{}
	n.URL = os.Getenv("NURSECALL_ENDPOINT")
	if len(n.URL) <= 0 {
		n.URL = "https://nursecall.io/api/v1/progresses"
	}

	n.CallToken = os.Getenv("NURSECALL_CALL_TOKEN")
	n.Debug = "TRUE" == os.Getenv("NURSECALL_DEBUG")

	n.Start()

	exitStatus, err := do(cmd)
	if err == nil {
		if exitStatus == 0 {
			n.Done(exitStatus)
		} else {
			n.Error(strconv.Itoa(int(exitStatus)))
		}
	} else {
		n.Error(err.Error())
	}
}

func do(cmd *exec.Cmd) (int, error) {

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("error: cmd.StdoutPipe(): ", err)
		return 0, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Println("error: cmd.StderrPipe(): ", err)
		return 0, err
	}

	defer stdoutPipe.Close()
	defer stderrPipe.Close()

	err = cmd.Start()
	if err != nil {
		log.Println("error: cmd.Start(): ", err)
		return 0, err
	}

	go printOutput(stdoutPipe)
	go printOutput(stderrPipe)

	exitStatus := 0
	if err := cmd.Wait(); err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			} else {
				//log.Println(err)
			}
		}
	}

	return exitStatus, err
}

func printOutput(reader io.Reader) {
	var (
		err error
		n   int
	)
	buf := make([]byte, 1024)

	for {
		if n, err = reader.Read(buf); err != nil {
			break
		}

		fmt.Println(string(buf[0:n]))
	}

	if err != io.EOF {
		//log.Println("error: err != io.EOF: " + err.Error())
	}
}

