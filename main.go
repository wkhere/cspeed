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

	help func()
}

func defaults() pargs { return pargs{dt: 5 * time.Second} }

func parseArgs(args []string) (a pargs, err error) {
	const usage = "usage: cspeed [-dDT|--duration=DT] HOST"

	a = defaults()

	rest := make([]string, 0, len(args))
	var p pstate

flags:
	for ; len(args) > 0 && p.err == nil; args = args[1:] {
		switch arg := args[0]; {

		case p.parseDurationFlag(arg, "-d", "--duration", &a.dt):
			// ok

		case arg == "-h", arg == "--help":
			a.help = func() { fmt.Println(usage) }
			return a, nil

		case arg == "--":
			rest = append(rest, args[1:]...)
			break flags

		case len(arg) > 1 && arg[0] == '-':
			p.errorf("unknown flag %s", arg)

		default:
			rest = append(rest, arg)
		}
	}

	if p.err != nil {
		return a, p.err
	}

	if len(rest) != 1 {
		return a, fmt.Errorf("expecting HOST arg")
	}
	a.host = rest[0]
	return a, nil
}

func main() {
	log.SetFlags(0)

	a, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}
	if a.help != nil {
		a.help()
		os.Exit(0)
	}

	x, err := sshSend(a.host, a.dt)
	if err != nil {
		die(1, err)
	}
	fmt.Println(x/1024, "KBps")
}

func die(exitcode int, err error) {
	log.Println(err)
	os.Exit(exitcode)
}
