package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type pargs struct {
	host string
	dt   time.Duration
}

func defaults() pargs { return pargs{dt: 5 * time.Second} }

func run(p *pargs) (nBps int64, err error) {
	c1 := exec.Command("dd", "if=/dev/urandom")
	c2 := exec.Command("ssh", p.host, "cat - >/dev/null")

	pr, pw := io.Pipe()
	var (
		diag   = new(strings.Builder)
		sshErr = new(strings.Builder)
	)
	c1.Stdout, c2.Stdin = pw, pr
	c1.Stderr = diag
	c2.Stderr = sshErr

	err = c1.Start()
	if err != nil {
		return 0, err
	}
	err = c2.Start()
	if err != nil {
		return 0, err
	}

	go func() {
		c1.Wait()
		pw.Close()
	}()

	done := make(chan *os.ProcessState)

	go func() {
		c2.Wait()
		done <- c2.ProcessState
		close(done)
	}()

	var (
		stop     = time.NewTimer(p.dt)
		graceEnd = make(chan struct{})
	)
loop:
	for {
		select {
		case <-stop.C:
			c1.Process.Signal(os.Interrupt)
			time.AfterFunc(2000*time.Millisecond, func() {
				close(graceEnd)
			})

		case p := <-done:
			if !p.Success() {
				return 0, fmt.Errorf(strings.TrimRight(sshErr.String(), "\n\r"))
			}
			break loop

		case <-graceEnd:
			c2.Process.Kill()
			return 0, fmt.Errorf("ssh: timeout")
		}
	}

	m := rxBps.FindStringSubmatch(diag.String())
	if len(m) < 2 {
		return 0, errNoBps
	}
	return strconv.ParseInt(m[1], 10, 64)
}

var (
	rxBps    = regexp.MustCompile(`\((\d+) bytes\/sec\)`)
	errNoBps = fmt.Errorf("Bps diagnostics not found")
)

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

	x, err := run(&p)
	if err != nil {
		die(1, err)
	}
	fmt.Println("KBps:", x/1024)
}

func die(exitcode int, err error) {
	log.Println(err)
	os.Exit(exitcode)
}
