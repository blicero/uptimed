// /home/krylon/go/src/github.com/blicero/uptimed/client/data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 10. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-10 17:57:19 krylon>

package client

import (
	"time"

	"github.com/blicero/uptimed/common"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

// getData gets the current system uptime and load average.
func (c *Client) getData() (*common.Record, error) {
	var (
		err        error
		uptimeSecs uint64
		loadavg    *load.AvgStat
		r          = &common.Record{
			Hostname:  c.name,
			Timestamp: time.Now(),
		}
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
