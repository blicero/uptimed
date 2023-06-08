// /home/krylon/go/src/github.com/blicero/uptimed/dnssd/service.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-07 23:56:47 krylon>

package dnssd

import (
	"log"

	"github.com/blicero/uptimed/common"
	"github.com/blicero/uptimed/logdomain"
	"github.com/hashicorp/mdns"
)

type Service struct {
	hostname string
	port     int
	log      *log.Logger
	svc      *mdns.MDNSService
	srv      *mdns.Server
}

func CreateService(hostname string, port int) (*mdns.MDNSService, error) {
	var (
		err error
		s   = &Service{hostname: hostname, port: port}
	)

	if s.log, err = common.GetLogger(logdomain.DNSSD); err != nil {
		return nil, err
	} else if s.svc, err = mdns.NewMDNSService("uptimed", "_http._tcp", "", hostname, port, nil, []string{"Abobo"}); err != nil {
		s.log.Printf("[ERROR] Cannot register service on MDNS: %s\n",
			err.Error())
		return nil, err
	}
} // func CreateService(addr string, port int) (*mdns.MDNSService, error)
