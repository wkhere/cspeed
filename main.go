package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/wkhere/argp"
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
		"usage: cspeed [-dDT|--duration=DT (default %[1]s)] HOST [HOSTn...]\n",
		a.dt,
	)
}

func parseArgs(args []string) (a pargs, err error) {
	a = defaults()

	var rest []string
	var p argp.PState
flags:
	for ; len(args) > 0; args = args[1:] {
		switch arg := args[0]; {

		case p.IsFlagExpr(arg, "-d", "--duration"):
			p.ParseDuration(&a.dt)

		case arg == "-h", arg == "--help":
			a.help = usage
			return a, nil

		case arg == "--":
			rest = append(rest, args[1:]...)
			break flags

		case len(arg) > 1 && arg[0] == '-':
			p.Errorf("unknown flag %s", arg)

		default:
			rest = append(rest, arg)
		}
	}

	switch {
	case p.Err != nil:
		return a, p.Err

	case len(rest) < 1:
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

	var maxHostLen int
	for _, host := range a.hosts {
		maxHostLen = max(maxHostLen, len(host))
	}
	printHeader := func(host string) { fmt.Printf("\r%-*s\t", maxHostLen+1, host+":") }

	var hadErr bool
	for _, host := range a.hosts {
		printHeader(host)

		done := make(chan struct{})
		sync := make(chan struct{})

		go func() {
			t0 := time.Now()
			ticker := time.NewTicker(time.Second)
			for {
				select {
				case t1 := <-ticker.C:
					printHeader(host)
					fmt.Printf("%5s  ", t1.Sub(t0).Truncate(time.Second))
				case <-done:
					close(sync)
					return
				}
			}
		}()

		x, err := sshSend(host, a.dt)
		close(done)
		<-sync

		if err != nil {
			hadErr = true
			fmt.Println()
			log.Printf("%s: %s", host, err)
			continue
		}

		printHeader(host)
		fmt.Printf("%4d KBps   \n", x/1024)

		time.Sleep(100 * time.Millisecond)
	}

	if hadErr {
		os.Exit(1)
	}
}
