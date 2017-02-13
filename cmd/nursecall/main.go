package main

import (
	"log"
	"os"

	"github.com/narita-takeru/nursecall"
)

func main() {
	if len(os.Args) == 0 {
		log.Println("No args")
		return
	}
	cmdStr := os.Args[1]
	args := os.Args[2:]
	tokens := os.Args[1:]
	if err := nursecall.Start(cmdStr, args, nursecall.NewNotifier(tokens)); err != nil {
		log.Println(err)
	}
}
