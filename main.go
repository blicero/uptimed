// /home/krylon/go/src/github.com/blicero/uptimed/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-12 18:42:56 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/blicero/uptimed/client"
	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/web"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	const defaultAddr = "[::1]"

	var (
		err        error
		mode, addr string
		port       int64
		mdns       bool
	)

	switch runtime.GOOS {
	case "freebsd", "openbsd":
		mdns = false
	default:
		mdns = true
	}

	flag.StringVar(&mode, "mode", "client", "Tells if we are a client or a server")
	flag.StringVar(&addr, "addr", defaultAddr, "The network address to listen on (server) or report to (client)")
	flag.Int64Var(&port, "port", common.WebPort, "The TCP port for the HTTP server to listen on")
	flag.BoolVar(&mdns, "mdns", mdns, "Use mDNS for server discovery")

	flag.Parse()

	if err = common.InitApp(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Cannot initialize directory %s: %s\n",
			common.BaseDir,
			err.Error(),
		)
		os.Exit(1)
	}

	addr = fmt.Sprintf("[%s]:%d",
		addr,
		port)

	switch mode {
	case "client":
		// Do something!
		var c *client.Client

		if c, err = client.Create(addr, mdns); err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to create client: %s\n",
				err.Error())
			os.Exit(1)
		}

		c.Loop()
	case "server":
		var srv *web.Server

		if srv, err = web.Open(addr, int(port)); err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to instantiate web server: %s\n",
				err.Error())
			os.Exit(1)
		}

		go srv.ListenAndServe()

	}

	var sigQ = make(chan os.Signal, 1)

	signal.Notify(sigQ, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	sig := <-sigQ
	fmt.Printf("Quitting on signal %s\n", sig)

	os.Exit(0)
}
