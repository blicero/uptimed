// /home/krylon/go/src/github.com/blicero/uptimed/dnssd/01_srv_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-09 14:28:50 krylon>

package dnssd

import (
	"testing"

	"github.com/blicero/uptimed/common"
)

var srv *Server

func TestServerCreate(t *testing.T) {
	var err error

	if srv, err = CreateService("schwarzgeraet.", int(common.WebPort)-1); err != nil {
		srv = nil
		t.Fatalf("Cannot create Server: %s\n",
			err.Error())
	}

	//time.Sleep(time.Second * 30)

} // func TestServerCreate(t *testing.T)
