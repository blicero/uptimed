// /home/krylon/go/src/github.com/blicero/uptimed/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-01 18:44:03 krylon>

package database

import "github.com/blicero/uptimed/database/query"

var qDB = map[query.ID]string{
	query.HostGetID:  "SELECT id FROM host WHERE name = ?",
	query.HostGetAll: "SELECT id, name FROM host",
	query.HostAdd:    "INSERT INTO host (name) VALUES (?) RETURNING id",
	query.RecordAdd: `
INSERT INTO record (host_id, timestamp, uptime, load1, load5, load15)
VALUES             (      ?,         ?,      ?,     ?,     ?,      ?)
`,
	query.RecordGetByPeriod: `
SELECT
  r.id,
  h.name,
  r.timestamp,
  r.uptime,
  r.load1,
  r.load5,
  r.load15
FROM record r
JOIN host h ON (r.host_id = h.id)
WHERE r.timestamp BETWEEN ? AND ?
`,
	query.RecordGetByHost: `
SELECT
  id,
  timestamp,
  uptime,
  load1,
  load5,
  load15
FROM record
WHERE host_id = ?
`,
}
