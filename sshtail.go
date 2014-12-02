package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n\n  %s comma,separated,hostnames files to tail\n\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

func fatalf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(2)
}

func main() {
	username := os.Getenv("USER")
	cmd := "tail -F"

	flag.StringVar(&username, "u", username, "username")
	flag.StringVar(&cmd, "cmd", cmd, "command to run (files are appended)")

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		usage()
		os.Exit(1)
	}

	hosts := []string{}
	for _, host := range strings.Split(args[0], ",") {
		if !strings.Contains(host, ":") {
			host += ":22"
		}
		hosts = append(hosts, host)
	}
	for _, f := range args[1:] {
		cmd += " " + f
	}

	conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		fatalf("error connecting to ssh-agent: %v", err)
	}
	defer conn.Close()
	ag := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)},
	}

	wg := sync.WaitGroup{}
	for _, host := range hosts {
		wg.Add(1)
		go tail(cmd, host, config, wg.Done)
	}
	wg.Wait()
}
