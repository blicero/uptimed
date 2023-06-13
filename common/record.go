// /home/krylon/go/src/github.com/blicero/uptimed/common/data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-13 19:18:27 krylon>

package common

import (
	"fmt"
	"time"
)

// Host is a system - physical or virtual - that is connected to a network.
type Host struct {
	ID   int64
	Name string
}

func (h *Host) String() string {
	return fmt.Sprintf("Host{ ID: %d, Name: %s }",
		h.ID,
		h.Name)
}

// Record is a data record the client submits to the server.
type Record struct {
	ID        int64
	Hostname  string
	Timestamp time.Time
	Uptime    time.Duration
	Load      [3]float64
}

// Recent returns true of the Record was submitted recently (now - interval * 2)
func (r *Record) Recent() bool {
	return time.Since(r.Timestamp) < Interval*2
} // func (r *Record) Recent() bool

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
