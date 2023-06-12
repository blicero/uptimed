// /home/krylon/go/src/github.com/blicero/uptimed/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-12 23:13:38 krylon>

// Package client implements the data acquisition and communication with
// the server.
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/dnssd"
	"github.com/blicero/uptimed/logdomain"
)

// Client contains the state for the client side, i.e. for data acquisition and
// communication with the server.
type Client struct {
	srvAddr     string
	origSrvAddr string
	hc          http.Client
	log         *log.Logger
	name        string
	res         *dnssd.Resolver
	alive       atomic.Bool
}

const reqPath = "/ws/report"

// Create creates a new Client.
func Create(addr string, mdns bool) (*Client, error) {
	var (
		err     error
		addrStr string
		paddr   *url.URL
		c       = &Client{
			hc:      http.Client{Transport: &http.Transport{DisableCompression: false}},
			srvAddr: addr,
		}
	)

	if c.name, err = os.Hostname(); err != nil {
		fmt.Printf("[ERROR] Cannot query hostname: %s\n",
			err.Error())
		return nil, err
	} else if i := strings.Index(c.name, "."); i != -1 {
		c.name = c.name[:i]
	} else if c.log, err = common.GetLogger(logdomain.Client); err != nil {
		return nil, err
	}

	if c.log == nil {
		if c.log, err = common.GetLogger(logdomain.Client); err != nil {
			fmt.Printf("On second attempt, common.GetLogger returned an error: %s\n",
				err.Error())
			return nil, err
		} else if c.log == nil {
			fmt.Printf("Logger is nil! Why?\n")
			return nil, errors.New("logger is nil")
		}
	}

	if mdns {
		c.log.Println("[DEBUG] Start mDNS service discovery")
		if c.res, err = dnssd.NewResolver(c.name); err != nil {
			c.log.Printf("[ERROR] Cannot initiate DNS-SD resolver: %s\n",
				err.Error())
			return nil, err
		}

		go c.res.FindServer()
	} else {
		c.log.Println("[DEBUG] Disabling mDNS service discovery")
	}

	addrStr = fmt.Sprintf("http://%s%s",
		addr,
		reqPath)

	if paddr, err = url.Parse(addrStr); err != nil {
		return nil, err
	}

	c.srvAddr = paddr.String()
	c.origSrvAddr = c.srvAddr

	c.log.Printf("[DEBUG] Client %s initialized\n",
		c.name)

	return c, nil
} // func Create() (*Client, error)

// Alive returns the value of the Client's alive flag
func (c *Client) Alive() bool {
	return c.alive.Load()
} // func (c *Client) Alive() bool

// Stop tells the Client to cease activity.
func (c *Client) Stop() {
	c.alive.Store(false)
} // func (c *Client) Stop()

// Loop executes the Client's main loop.
func (c *Client) Loop() {
	c.alive.Store(true)

	go c.gatherLoop()
	go c.transmitLoop()
} // func (c *Client) Loop() error

func (c *Client) gatherLoop() {
	var t = time.NewTicker(common.Interval)
	defer t.Stop()

	c.log.Printf("[TRACE] Start gathering data on %s\n",
		c.name)

	for c.Alive() {
		<-t.C

		var (
			err error
			rec *common.Record
			buf []byte
		)

		c.log.Println("[TRACE] Gathering data")

		if rec, err = c.getData(); err != nil {
			c.log.Printf("[ERROR] Cannot acquire data: %s\n", err.Error())
		} else if buf, err = json.Marshal(&rec); err != nil {
			c.log.Printf("[ERROR] Cannot serialize data: %s\n", err.Error())
		} else if err = c.saveBuffer(buf); err != nil {
			c.log.Printf("[ERROR] Cannot save data: %s\n", err.Error())
		}
	}
} // func (c *Client) gatherLoop()

func (c *Client) transmitLoop() {
	var t = time.NewTicker(common.Interval)
	defer t.Stop()

	c.log.Printf("[TRACE] Start transmitting buffered data on %s\n",
		c.name)

	for c.Alive() {
		var (
			err   error
			files []string
			glob  string
		)

		<-t.C

		c.log.Printf("[TRACE] Transmitting buffered data to %s\n",
			c.srvAddr)

		glob = filepath.Join(common.BufferPath, "*.json")

		if files, err = filepath.Glob(glob); err != nil {
			c.log.Printf("[ERROR] Cannot lookup files in %s: %s\n",
				common.BufferPath,
				err.Error())
			continue
		}

		for _, f := range files {
			var buf []byte

			if buf, err = c.slurp(f); err != nil {
				c.log.Printf("[ERROR] Cannot slurp file %s: %s\n",
					f,
					err.Error())
				continue
			} else if err = c.transmitData(buf); err != nil {
				c.log.Printf("[ERROR] Failed to send data to %s: %s\n",
					c.srvAddr,
					err.Error())
				continue
			} else if err = os.RemoveAll(f); err != nil {
				c.log.Printf("[ERROR] Cannot remove %s: %s\n",
					f,
					err.Error())
			} else {
				c.log.Printf("[TRACE] Processed %s\n", f)
			}
		}
	}
} // func (c *Client) transmitLoop()

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

func (c *Client) slurp(path string) ([]byte, error) {
	var (
		err error
		fh  *os.File
		buf bytes.Buffer
	)

	if fh, err = os.Open(path); err != nil {
		c.log.Printf("[ERROR] Cannot open %s: %s\n",
			path,
			err.Error())
		return nil, err
	}

	defer fh.Close()

	if _, err = io.Copy(&buf, fh); err != nil {
		c.log.Printf("[ERROR] Cannot read file %s: %s\n",
			path,
			err.Error())
		return nil, err
	}

	return buf.Bytes(), nil
} // func (c *Client) slurp(path string) ([]byte, error)
