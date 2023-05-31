// /home/krylon/go/src/github.com/blicero/uptimed/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-05-31 18:16:13 krylon>

// Package client implements the data acquisition and communication with
// the server.
package client

import (
	"log"
	"time"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/logdomain"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

// Client contains the state for the client side, i.e. for data acquisition and
// communication with the server.
type Client struct {
	log *log.Logger
}

// Create creates a new Client.
func Create() (*Client, error) {
	var (
		err error
		c   = new(Client)
	)

	if c.log, err = common.GetLogger(logdomain.Client); err != nil {
		return nil, err
	}

	return c, nil
} // func Create() (*Client, error)

// GetData gets the current system uptime and load average.
func (c *Client) GetData() (*common.Record, error) {
	var (
		err        error
		uptimeSecs uint64
		loadavg    *load.AvgStat
		r          = &common.Record{Timestamp: time.Now()}
	)

	if uptimeSecs, err = host.Uptime(); err != nil {
		c.log.Printf("[ERROR] Cannot determine uptime: %s\n",
			err.Error())
		return nil, err
	}

	r.Uptime = time.Second * time.Duration(uptimeSecs)

	if loadavg, err = load.Avg(); err != nil {
		c.log.Printf("[ERROR] Cannot query system load: %s\n",
			err.Error())
		return nil, err
	}

	r.Load[0] = loadavg.Load1
	r.Load[1] = loadavg.Load5
	r.Load[2] = loadavg.Load15

	return r, nil
} // func (c *Client) GetData() (*common.Record, error)
