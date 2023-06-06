// /home/krylon/go/src/github.com/blicero/uptimed/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-06 18:45:59 krylon>

// Package client implements the data acquisition and communication with
// the server.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/logdomain"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

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
		)

		c.processBuffered()

		if data, err = c.GetData(); err != nil {
			c.log.Printf("[ERROR] Cannot acquire data: %s\n", err.Error())
			goto NEXT
		} else if buf, err = json.Marshal(data); err != nil {
			c.log.Printf("[ERROR] Cannot serialize data: %s\n", err.Error())
			goto NEXT
		} else if err = c.transmitData(buf); err != nil {
			c.log.Printf("[ERROR] Cannot send data to server: %s\n",
				err.Error())
			c.saveBuffer(buf) // nolint: errcheck
		}

		c.log.Printf("[INFO] Report sent to Server %s\n",
			c.srvAddr)

	NEXT:
		time.Sleep(interval)
	}
} // func (c *Client) Run() error

func (c *Client) transmitData(buf []byte) error {
	var (
		err error
		rdr = bytes.NewReader(buf)
		res *http.Response
	)

	if res, err = c.hc.Post(c.srvAddr, common.EncJSON, rdr); err != nil {
		c.log.Printf("[ERROR] Cannot send data to server %s: %s\n",
			c.srvAddr,
			err.Error())

		return err
	}

	res.Body.Close()

	return nil
}

func (c *Client) saveBuffer(buf []byte) error {
	var (
		err            error
		filename, path string
		fh             *os.File
	)

	filename = time.Now().Format("20060102-150405.json")
	path = filepath.Join(common.BufferPath, filename)

	if fh, err = os.Create(path); err != nil {
		c.log.Printf("[ERROR] Cannot open %s: %s\n",
			path,
			err.Error())

		return err
	}

	defer fh.Close() // nolint: errcheck

	if _, err = fh.Write(buf); err != nil {
		c.log.Printf("[ERROR] Cannot write JSON data to %s: %s\n",
			path,
			err.Error())
		os.RemoveAll(path) // nolint: errcheck

		return err
	}

	return nil
} // func (c *Client) saveBuffer(buf []byte) error

func (c *Client) sendBufferedData(path string, wg *sync.WaitGroup) {
	var (
		err error
		fh  *os.File
		buf bytes.Buffer
	)

	defer wg.Done()

	if fh, err = os.Open(path); err != nil {
		c.log.Printf("[ERROR] Cannot open %s: %s\n",
			path,
			err.Error())
		return
	}

	defer fh.Close() // nolint: errcheck

	if _, err = io.Copy(&buf, fh); err != nil {
		c.log.Printf("[ERROR] Failed to read contents of %s: %s\n",
			path,
			err.Error())
		return
	} else if err = c.transmitData(buf.Bytes()); err != nil {
		c.log.Printf("[ERROR] Failed to transmit contents of %s to server: %s\n",
			path,
			err.Error())
		return
	}

	os.RemoveAll(path) // nolint: errcheck
} // func (c *Client) sendBufferedData(path string)

func (c *Client) processBuffered() {
	var (
		err   error
		files []string
		glob  = filepath.Join(common.BufferPath, "*.json")
	)

	if files, err = filepath.Glob(glob); err != nil {
		c.log.Printf("[ERROR] Cannot read filenames from %s: %s\n",
			common.BufferPath,
			err.Error())
		return
	} else if len(files) == 0 {
		return
	}

	var wg sync.WaitGroup

	for _, path := range files {
		wg.Add(1)
		go c.sendBufferedData(path, &wg)
	}

	wg.Wait()
} // func (c *Client) processBuffered()
