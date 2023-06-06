// /home/krylon/go/src/github.com/blicero/uptimed/web/tmpl_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-06 23:43:27 krylon>

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
	Clients   []common.Host
}

type tmplDataMain struct {
	tmplDataBase
	Records []common.Record
}

type tmplDataHost struct {
	tmplDataBase
	Hostname string
	Records  []common.Record
	Period   int64
}

type tmplDataPrefs struct {
	tmplDataBase
	Period int64
}
