// /home/krylon/go/src/github.com/blicero/uptimed/dnssd/02_res_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-09 18:46:26 krylon>

package dnssd

import (
	"os"
	"testing"
	"time"
)

var res *Resolver

func TestCreateResolver(t *testing.T) {
	var (
		err      error
		hostname string
	)

	if hostname, err = os.Hostname(); err != nil {
		t.Fatalf("Cannot query hostname from OS: %s", err.Error())
	} else if res, err = NewResolver(hostname); err != nil {
		res = nil
		t.Fatalf("Cannot create Resolver: %s", err.Error())
	}
} // func TestCreateResolver(t *testing.T)

func TestFindServer(t *testing.T) {
	if res == nil {
		t.SkipNow()
	}

	go find(res)

	time.Sleep(time.Second * 2)

	res.pLock.RLock()
	defer res.pLock.RUnlock()

	for name, svc := range res.servers {
		t.Logf("%s => %#v\n",
			name,
			svc)
	}
} // func TestFindServer()

func TestStopResolver(t *testing.T) {
	if res == nil {
		t.SkipNow()
	}

	res.Stop()
} // func TestStopResolver(t *testing.T)

func find(r *Resolver) {
	r.FindServer()
} // func find(r *Resolver)
