// /home/krylon/go/src/github.com/blicero/uptimed/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-04 17:37:02 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
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

	var (
		err         error
		mode, addr  string
		defaultAddr = fmt.Sprintf("[::1]:%d", common.WebPort)
	)

	flag.StringVar(&mode, "mode", "client", "Tells if we are a client or a server")
	flag.StringVar(&addr, "addr", defaultAddr, "The network address to listen on (server) or report to (client)")

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

	switch mode {
	case "client":
		// Do something!
		var c *client.Client

		if c, err = client.Create(addr); err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to create client: %s\n",
				err.Error())
			os.Exit(1)
		}

		go c.Run()
	case "server":
		var srv *web.Server

		if srv, err = web.Open(addr); err != nil {
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
