package main

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func sshSend(host string, dt time.Duration) (nBps int64, err error) {

	c1 := exec.Command("dd", "if=/dev/urandom")
	c2 := exec.Command("ssh", host, "cat - >/dev/null")

	pr, pw, err := os.Pipe()
	if err != nil {
		return 0, err
	}
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

	done := make(chan error)

	go func() {
		done <- c2.Wait()
		close(done)
	}()

	var (
		stop     = time.NewTimer(dt)
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

		case err := <-done:
			if err != nil {
				// todo: handle case of ssh external kill (sshErr is empty then)
				return 0, errSSHProc{strings.TrimRight(sshErr.String(), "\n\r")}
			}
			break loop

		case <-graceEnd:
			c2.Process.Kill()
			return 0, errSSHTimeout
		}
	}

	m := rxBps.FindStringSubmatch(diag.String())
	if len(m) < 2 {
		return 0, errNoBps
	}
	return strconv.ParseInt(m[1], 10, 64)
}

var (
	rxBps = regexp.MustCompile(`\((\d+) bytes\/sec\)`)

	errNoBps      = errors.New("Bps diagnostics not found")
	errSSHTimeout = errors.New("ssh: timeout")
)

type errSSHProc struct{ string }

func (e errSSHProc) Error() string { return e.string }
