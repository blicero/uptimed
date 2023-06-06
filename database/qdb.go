// /home/krylon/go/src/github.com/blicero/uptimed/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-06 23:13:19 krylon>

package database

import "github.com/blicero/uptimed/database/query"

var qDB = map[query.ID]string{
	query.HostGetID:  "SELECT id FROM host WHERE name = ?",
	query.HostGetAll: "SELECT id, name FROM host ORDER BY name",
	query.HostAdd:    "INSERT INTO host (name) VALUES (?) RETURNING id",
	query.RecordAdd: `
INSERT INTO record (host_id, timestamp, uptime, load1, load5, load15)
VALUES             (      ?,         ?,      ?,     ?,     ?,      ?)
RETURNING id
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
ORDER BY r.timestamp
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
WHERE host_id = ? AND timestamp >= ?
ORDER BY timestamp
`,
	query.RecentGetAll: `
WITH data AS (
SELECT
        r.id,
        h.name,
        row_number() OVER (PARTITION BY h.name ORDER BY r.timestamp DESC) AS hid,
        -- datetime(r.timestamp, 'unixepoch', 'localtime') AS timestamp,
        r.timestamp,
        r.uptime,
        r.load1,
        r.load5,
        r.load15
FROM host h
RIGHT JOIN record r ON h.id = r.host_id
)
SELECT id, name, timestamp, uptime, load1, load5, load15
FROM data
WHERE hid = 1
ORDER BY name
;
`,
}
