package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type pargs struct {
	host string
	dt   time.Duration
}

func defaults() pargs { return pargs{dt: 5 * time.Second} }

func parseArgs(args []string) (p pargs, _ error) {
	const usage = "usage: cspeed HOST"

	p = defaults()

	if len(args) != 1 {
		return p, fmt.Errorf(usage)
	}

	p.host = args[0]
	return p, nil
}

func main() {
	log.SetFlags(0)

	p, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}

	x, err := sshSend(&p)
	if err != nil {
		die(1, err)
	}
	fmt.Println(x/1024, "KBps")
}

func die(exitcode int, err error) {
	log.Println(err)
	os.Exit(exitcode)
}
