// /home/krylon/go/src/github.com/blicero/uptimed/web/ajax_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-13 19:48:42 krylon>

package web

import (
	"time"

	"github.com/blicero/uptimed/common"
)

type response struct {
	Status    bool
	Message   string
	Timestamp time.Time
}

type responseRecords struct {
	Status    bool                     `json:"status"`
	Message   string                   `json:"message"`
	Timestamp time.Time                `json:"timestamp"`
	Records   map[string]common.Record `json:"records"`
}
