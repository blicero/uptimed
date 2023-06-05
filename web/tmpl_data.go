// /home/krylon/go/src/github.com/blicero/uptimed/web/tmpl_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-05 20:55:22 krylon>

package web

import (
	"time"

	"github.com/blicero/uptimed/common"
)

// nolint: unused
type tmplDataBase struct {
	Title     string
	Timestamp time.Time
	Debug     bool
	URL       string
}

type tmplDataMain struct {
	tmplDataBase
	Clients []common.Host
	Records []common.Record
}

type tmplDataHost struct {
	tmplDataBase
	Hostname string
	Records  []common.Record
}
