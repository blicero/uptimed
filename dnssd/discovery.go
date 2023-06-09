// /home/krylon/go/src/github.com/blicero/uptimed/dnssd/discovery.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-09 18:56:18 krylon>

package dnssd

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/logdomain"
	"github.com/grandcat/zeroconf"
)

// Resolver uses mDNS to find server instances.
type Resolver struct {
	log          *log.Logger
	hostname     string
	res          *zeroconf.Resolver
	alive        atomic.Bool
	pLock        sync.RWMutex
	purgeRunning atomic.Bool
	servers      map[string]service
}

// NewResolver creates a new mDNS Resolver
func NewResolver(hostname string) (*Resolver, error) {
	var (
		err error
		r   = &Resolver{
			hostname: hostname,
			servers:  make(map[string]service),
		}
	)

	if r.log, err = common.GetLogger(logdomain.DNSSD); err != nil {
		return nil, err
	} else if r.res, err = zeroconf.NewResolver(nil); err != nil {
		r.log.Printf("[ERROR] Cannot create ZeroConf resolver: %s\n",
			err.Error())
		return nil, err
	}

	r.alive.Store(true)

	return r, nil
} // func NewResolver() (*Resolver, error)

// Stop stops the Resolver
func (r *Resolver) Stop() {
	r.alive.Store(false)
} // func (r *Resolver) Stop()

// Alive returns the alive flag of the resolver
func (r *Resolver) Alive() bool {
	return r.alive.Load()
} // func (r *Resolver) Alive() bool

// FindServer starts the Resolver's discovery process.
func (r *Resolver) FindServer() {
	go r.purgeLoop()

	for r.Alive() {
		var (
			err     error
			ctx     context.Context
			cancel  context.CancelFunc
			entries = make(chan *zeroconf.ServiceEntry)
		)

		// defer close(entries)

		go r.processServiceEntries(entries)

		ctx, cancel = context.WithCancel(context.Background())

		if err = r.res.Browse(ctx, srvService, srvDomain, entries); err != nil {
			r.log.Printf("{ERROR] Failed to browse for %s: %s\n",
				srvService,
				err.Error())
		}

		time.Sleep(time.Second * srvTTL)
		cancel()
		close(entries)

	}
} // func (r *Resolver) FindServer() ([]string, error)

func (r *Resolver) processServiceEntries(queue <-chan *zeroconf.ServiceEntry) {
	r.log.Println("[DEBUG] Waiting for records from mdns")

	for entry := range queue {
		var str = rrStr(entry)

		r.log.Printf("[DEBUG] Got one record: %s\n", str)

		if !peerPat.MatchString(entry.Instance) {
			continue
		}

		entry.TTL = srvTTL

		r.pLock.Lock()
		r.servers[str] = mkService(entry)
		r.pLock.Unlock()
	}
} // func (d *Daemon) processServiceEntries(queue <- chan *zeroconf.ServiceEntry)

func (r *Resolver) purgeLoop() {
	if r.purgeRunning.Load() {
		return
	}

	r.purgeRunning.Store(true)
	defer r.purgeRunning.Store(false)

	for r.Alive() {
		time.Sleep(time.Second * srvTTL)

		r.pLock.Lock()
		for k, srv := range r.servers {
			if srv.isExpired() {
				r.log.Printf("[DEBUG] Remove Peer %s from cache\n",
					k)
				delete(r.servers, k)
			}
		}
		r.pLock.Unlock()
	}
} // func (d *Daemon) purgeLoop()
