// /home/krylon/go/src/github.com/blicero/uptimed/common/data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-05-31 19:01:41 krylon>

package common

import (
	"fmt"
	"time"
)

// Record is a data record the client submits to the server.
type Record struct {
	Hostname  string
	Timestamp time.Time
	Uptime    time.Duration
	Load      [3]float64
}

func (r *Record) String() string {
	return fmt.Sprintf("Record{ Hostname: %q, Timestamp: %s, Uptime: %s, Load: { %.1f, %.1f, %.1f } }",
		r.Hostname,
		r.Timestamp.Format(TimestampFormat),
		r.Uptime,
		r.Load[0],
		r.Load[1],
		r.Load[2],
	)
}
