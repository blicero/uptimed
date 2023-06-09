// /home/krylon/go/src/github.com/blicero/uptimed/client/01_client_data_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 05. 2023 by Benjamin Walkenhorst
// (c) 2023 Benjamin Walkenhorst
// Time-stamp: <2023-06-12 18:08:32 krylon>

package client

import (
	"testing"

	"github.com/blicero/uptimed/common"
)

var tc *Client

func TestClientCreate(t *testing.T) {
	var err error

	if tc, err = Create("", false); err != nil {
		tc = nil
		t.Errorf("Cannot create Client instance: %s",
			err.Error())
	}
} // func TestClientCreate(t *testing.T)

func TestClientData(t *testing.T) {
	var (
		err    error
		record *common.Record
	)

	if tc == nil {
		t.SkipNow()
	}

	if record, err = tc.getData(); err != nil {
		t.Fatalf("Cannot acquire data: %s",
			err.Error())
	} else if record == nil {
		t.Error("GetData did not return an error, but it didn't return any data, either.")
	}

	t.Logf("Acquired data: %s", record)
} // func TestClientData(t *testing.T)
