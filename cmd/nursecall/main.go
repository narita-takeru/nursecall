package main

import (
	"log"
	"os"

	"github.com/narita-takeru/nursecall"
)

func main() {
	if len(os.Args) == 1 {
		log.Println("Error: no command provided. Usage: nursecall <command> [args...]")
		return
	}

	cmdStr := os.Args[1]
	args := os.Args[2:]
	if err := nursecall.Start(cmdStr, args, nursecall.NewNotifier()); err != nil {
		log.Println(err)
	}
}
