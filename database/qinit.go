// /home/krylon/go/src/github.com/blicero/uptimed/database/qinit.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-01 18:25:01 krylon>

package database

var qInit = []string{
	`
CREATE TABLE host (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
)
`,
	`
CREATE TABLE record (
    id INTEGER PRIMARY KEY,
    host_id INTEGER NOT NULL,
    timestamp INTEGER NOT NULL,
    uptime INTEGER NOT NULL,
    load1 REAL NOT NULL,
    load5 REAL NOT NULL,
    load15 REAL NOT NULL,
    FOREIGN KEY (host_id) REFERENCES host (id),
    UNIQUE (timestamp, host_id),
    CHECK (uptime >= 0),
    CHECK (load1 >= 0.0 AND load5 >= 0.0 AND load15 >= 0.0),
    CHECK (hostname <> '')
)
`,
	"CREATE INDEX rec_host_idx ON record (host_id)",
	"CREATE INDEX rec_stamp_idx ON record (timestamp)",
}
