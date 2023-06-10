// /home/krylon/go/src/github.com/blicero/uptimed/web/01_web_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-10 13:29:14 krylon>

package web

import (
	"fmt"
	"testing"

	"github.com/blicero/uptimed/common"
)

var (
	addr = fmt.Sprintf("[::1]:%d", common.WebPort+2)
	srv  *Server
)

func TestOpen(t *testing.T) {
	var err error

	if srv, err = Open(addr, int(common.WebPort)+2); err != nil {
		srv = nil
		t.Fatalf("Error creating web server: %s",
			err.Error())
	}
} // func TestOpen(t *testing.T)
