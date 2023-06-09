// /home/krylon/go/src/github.com/blicero/uptimed/dnssd/res_helper.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-09 18:49:45 krylon>

package dnssd

import (
	"fmt"
	"regexp"
	"time"

	"github.com/blicero/uptimed/common"
	"github.com/grandcat/zeroconf"
)

type service struct {
	rr        *zeroconf.ServiceEntry
	timestamp time.Time
}

// RR is a record of uptimed running on some machine on the local network
type RR struct {
	Instance string
	Hostname string
	Domain   string
	Port     int
}

var peerPat = regexp.MustCompile(fmt.Sprintf("%s\\\\@(\\w+)", common.AppName))

func (s *service) mkRR() RR { // nolint: unused
	return RR{
		Instance: s.rr.Instance,
		Hostname: s.rr.HostName,
		Domain:   s.rr.Domain,
		Port:     s.rr.Port,
	}
} // func (s *service) mkPeer() objects.Peer

func mkService(rr *zeroconf.ServiceEntry) service {
	return service{
		rr:        rr,
		timestamp: time.Now(),
	}
}

func rrStr(rr *zeroconf.ServiceEntry) string {
	return fmt.Sprintf("%s:%d",
		rr.HostName,
		rr.Port)
} // func rrStr(rr *zeroconf.ServiceEntry) string

func (s *service) isExpired() bool {
	return s.timestamp.Add(time.Second * time.Duration(s.rr.TTL)).Before(time.Now())
} // func (s *service) IsExpired() bool
