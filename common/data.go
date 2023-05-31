// /home/krylon/go/src/github.com/blicero/uptimed/common/data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-05-31 16:14:55 krylon>

package common

import "time"

// Record is a data record the client submits to the server.
type Record struct {
	Hostname  string
	Timestamp time.Time
	Uptime    time.Duration
	Load      [3]float64
}
