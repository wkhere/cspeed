package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type pargs struct {
	host string
	dt   time.Duration

	help func()
}

func defaults() pargs { return pargs{dt: 5 * time.Second} }

func parseArgs(args []string) (p pargs, err error) {
	const usage = "usage: cspeed [-d|--duration=DURATION] HOST"

	p = defaults()

	rest := make([]string, 0, len(args))
flags:
	for ; len(args) > 0; args = args[1:] {
		switch arg := args[0]; {

		case strings.HasPrefix(arg, "-d"), strings.HasPrefix(arg, "--duration"):
			flag, val, ok := strings.Cut(arg, "=")
			if !ok {
				return p, fmt.Errorf("expected %s=DURATION", arg)
			}
			p.dt, err = time.ParseDuration(val)
			if err != nil {
				return p, fmt.Errorf("flag %s: %w", flag, err)
			}

		case arg == "-h", arg == "--help":
			p.help = func() { fmt.Println(usage) }
			return p, nil

		case arg == "--":
			rest = append(rest, args[1:]...)
			break flags

		case len(arg) > 1 && arg[0] == '-':
			return p, fmt.Errorf("unknown flag %s", arg)

		default:
			rest = append(rest, arg)
		}
	}

	if len(rest) != 1 {
		return p, fmt.Errorf("expecting HOST arg")
	}
	p.host = rest[0]
	return p, nil
}

func main() {
	log.SetFlags(0)

	p, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}
	if p.help != nil {
		p.help()
		os.Exit(0)
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
