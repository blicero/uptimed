// /home/krylon/go/src/github.com/blicero/uptimed/web/ajax_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-03 15:25:34 krylon>

package web

import "time"

type response struct {
	Status    bool
	Message   string
	Timestamp time.Time
}
