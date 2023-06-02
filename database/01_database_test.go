// /home/krylon/go/src/github.com/blicero/uptimed/database/01_database_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 02. 06. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-02 16:53:24 krylon>

package database

import (
	"testing"

	"github.com/blicero/uptimed/common"
)

var db *Database

func TestOpen(t *testing.T) {
	var err error

	if db, err = Open(common.DbPath); err != nil {
		db = nil
		t.Fatalf("Cannot open database at %s: %s",
			common.DbPath,
			err.Error())
	}
} // func TestOpen(t *testing.T)

func TestQueries(t *testing.T) {
	if db == nil {
		t.SkipNow()
	}

	var err error

	for qid := range qDB {
		if _, err = db.getQuery(qid); err != nil {
			t.Errorf("Cannot prepare query %s: %s",
				qid,
				err.Error())
		}
	}
} // func TestQueries(t *testing.T)
