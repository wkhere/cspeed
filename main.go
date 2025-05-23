package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type pargs struct {
	hosts []string
	dt    time.Duration

	help func()
}

func defaults() pargs { return pargs{dt: 5 * time.Second} }

func usage() {
	a := defaults()
	fmt.Printf(
		"usage: cspeed [-dDT|--duration=DT (default %[1]s)] HOST [HOSTn...]",
		a.dt,
	)
}

func parseArgs(args []string) (a pargs, err error) {
	a = defaults()

	rest := make([]string, 0, len(args))
	var p pstate

flags:
	for ; len(args) > 0 && p.err == nil; args = args[1:] {
		switch arg := args[0]; {

		case p.parseDurationFlag(arg, "-d", "--duration", &a.dt):
			// ok

		case arg == "-h", arg == "--help":
			a.help = usage
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

	if len(rest) < 1 {
		return a, fmt.Errorf("expecting at least one HOST arg")
	}
	a.hosts = rest
	return a, nil
}

func main() {
	log.SetFlags(0)

	a, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	if a.help != nil {
		a.help()
		os.Exit(0)
	}

	var hadErr bool
	for _, host := range a.hosts {

		x, err := sshSend(host, a.dt)
		if err != nil {
			hadErr = true
			log.Printf("%s: %s", host, err)
			continue
		}
		fmt.Printf("%s: %d KBps\n", host, x/1024)

		time.Sleep(100 * time.Millisecond)
	}

	if hadErr {
		os.Exit(1)
	}
}
