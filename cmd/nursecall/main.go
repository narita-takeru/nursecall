package main

import (
	"os"
	"github.com/narita-takeru/nursecall"
)

func main() {
	cmdStr := os.Args[1]
	args := os.Args[2:]

	nursecall.Start(cmdStr, args)
}

