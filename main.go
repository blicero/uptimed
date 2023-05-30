// /home/krylon/go/src/github.com/blicero/uptimed/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-05-30 21:11:58 krylon>

package main

import (
	"fmt"
	"os"

	"github.com/blicero/uptimed/common"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	fmt.Fprintf(os.Stderr,
		"Implement me!\n")
}
