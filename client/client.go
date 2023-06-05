// /home/krylon/go/src/github.com/blicero/uptimed/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-05 19:20:49 krylon>

// Package client implements the data acquisition and communication with
// the server.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/logdomain"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

// FIXME: Increase to realistic value when done testing.
const interval = time.Second * 120

// Client contains the state for the client side, i.e. for data acquisition and
// communication with the server.
type Client struct {
	srvAddr string
	hc      http.Client
	log     *log.Logger
	name    string
}

const reqPath = "/ws/report"

// Create creates a new Client.
func Create(addr string) (*Client, error) {
	var (
		err     error
		addrStr string
		paddr   *url.URL
		c       = &Client{
			hc:      http.Client{Transport: &http.Transport{DisableCompression: false}},
			srvAddr: addr,
		}
	)

	addrStr = fmt.Sprintf("http://%s%s",
		addr,
		reqPath)

	if paddr, err = url.Parse(addrStr); err != nil {
		return nil, err
	}

	c.srvAddr = paddr.String()

	if c.log, err = common.GetLogger(logdomain.Client); err != nil {
		return nil, err
	} else if c.name, err = os.Hostname(); err != nil {
		c.log.Printf("[ERROR] Cannot query hostname: %s\n",
			err.Error())
		return nil, err
	} else if i := strings.Index(c.name, "."); i != -1 {
		c.name = c.name[:i]
	}

	c.log.Printf("[DEBUG] Client %s initialized\n",
		c.name)

	return c, nil
} // func Create() (*Client, error)

// GetData gets the current system uptime and load average.
func (c *Client) GetData() (*common.Record, error) {
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

// Run executes the Client's main loop.
func (c *Client) Run() error {
	for {
		var (
			err  error
			data *common.Record
			buf  []byte
			rdr  *bytes.Reader
		)

		if data, err = c.GetData(); err != nil {
			c.log.Printf("[ERROR] Cannot acquire data: %s\n", err.Error())
			goto NEXT
		} else if buf, err = json.Marshal(data); err != nil {
			c.log.Printf("[ERROR] Cannot serialize data: %s\n", err.Error())
			goto NEXT
		}

		rdr = bytes.NewReader(buf)

		if _, err = http.Post(c.srvAddr, common.EncJSON, rdr); err != nil {
			c.log.Printf("[ERROR] Cannot send data to server %s: %s\n",
				c.srvAddr,
				err.Error())
			goto NEXT
		}

		c.log.Printf("[INFO] Report sent to Server %s\n",
			c.srvAddr)

	NEXT:
		time.Sleep(interval)
	}
} // func (c *Client) Run() error
