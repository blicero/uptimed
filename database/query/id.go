// /home/krylon/go/src/github.com/blicero/uptimed/database/query/id.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-05 17:16:35 krylon>

// Package query provides symbolic constants to identify SQL queries to be
// run on the database.
package query

//go:generate stringer -type=ID

// ID identifies a database query.
type ID uint8

// These constants identify the queries that will be run on the database.
// The identifiers are meant to be self-explanatory.
const (
	HostGetID ID = iota
	HostGetAll
	HostAdd
	RecordAdd
	RecordGetByPeriod
	RecordGetByHost
	RecentGetAll
)
