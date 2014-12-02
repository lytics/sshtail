package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"
)

const (
	stdoutPipe = "\033[1;37m" + "out >>" + "\033[0m"
	stderrPipe = "\033[1;31m" + "err >>" + "\033[0m"
)

var newline = []byte{'\n'}

func tail(cmd, host string, config *ssh.ClientConfig, done func()) {
	defer func() {
		fmt.Fprintf(os.Stdout, "[%s] done\n", host)
		done()
	}()
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		fatalf("Failed to dial: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		fatalf("Failed to create stdout pipe: %v", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		fatalf("Failed to create stderr pipe: %v", err)
	}
	go pipe(host, stdoutPipe, stdout, os.Stdout)
	go pipe(host, stderrPipe, stderr, os.Stderr)
	if err := session.Run(cmd); err != nil {
		fatalf("Failed to run:", err)
	}
}

func pipe(host, name string, r io.Reader, w io.Writer) {
	prefix := fmt.Sprintf("%-10.10s %s ", host, name)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Fprintf(w, prefix)
		w.Write(scanner.Bytes())
		w.Write(newline)
	}
}
