// /home/krylon/go/src/github.com/blicero/uptimed/dnssd/service.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-12 17:03:41 krylon>

package dnssd

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/logdomain"
	"github.com/grandcat/zeroconf"
)

const (
	srvName    = "uptimed"
	srvService = "_http._tcp"
	srvDomain  = "local."
	srvTTL     = 90
)

// Server is a type.
type Server struct {
	log      *log.Logger
	srv      *zeroconf.Server
	hostname string
	port     int
	alive    atomic.Bool
}

// CreateService creates a new Server.
func CreateService(hostname string, port int) (*Server, error) {
	var (
		err error
		srv = &Server{
			hostname: hostname,
			port:     port,
		}
	)

	if srv.log, err = common.GetLogger(logdomain.DNSSD); err != nil {
		return nil, err
	}

	// FIXME
	// I copied this blindly from the example code without knowing
	// what it means or if it even has any significance at all.
	// For the time being, I do not have reliable Internet access,
	// so I cannot do any research on the matter.
	// But I suspect this is equivalent to "bla bla bla".
	var (
		instanceName = fmt.Sprintf("%s@%s",
			srvName,
			hostname)
		txt = []string{"txtv=0", "lo=1", "la=2"}
	)

	if srv.srv, err = zeroconf.Register(instanceName, srvService, srvDomain, port, txt, nil); err != nil {
		srv.log.Printf("[ERROR] Cannot register mDNS service: %s\n",
			err.Error())
		return nil, err
	}

	srv.alive.Store(true)

	return srv, nil
} // func CreateService(hostname string, port int) (*Server, error)

// IsAlive returns the Server's alive flag
func (srv *Server) IsAlive() bool {
	return srv.alive.Load()
} // func (srv.Server) IsAlive() bool

// Shutdown tells the Server to stop publishing the uptimed service
func (srv *Server) Shutdown() {
	if srv.alive.Load() {
		srv.srv.Shutdown()
		srv.alive.Store(false)
	}
} // func (srv *Server) Shutdown()
