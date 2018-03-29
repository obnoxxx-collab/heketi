//
// Copyright (c) 2018 The heketi Authors
//
// This file is licensed to you under your choice of the GNU Lesser
// General Public License, version 3 or any later version (LGPLv3 or
// later), or the GNU General Public License, version 2 (GPLv2), in all
// cases as published by the Free Software Foundation.
//

package glusterfs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/heketi/tests"
)

func TestCreateNodeHeathCache(t *testing.T) {
	tmpfile := tests.Tempfile()
	defer os.Remove(tmpfile)

	// Create the app (I'm being lazy here. An app is not strictly
	// needed but it is convenient.
	app := NewTestApp(tmpfile)
	defer app.Close()

	err := setupSampleDbWithTopology(app,
		2,    // clusters
		3,    // nodes_per_cluster
		4,    // devices_per_node,
		6*TB, // disksize)
	)
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	hc := NewNodeHealthCache(app.db, app.executor)
	tests.Assert(t, hc != nil, "expected hc != nil, got:", hc)

	nodeUp := hc.Status()
	tests.Assert(t, len(nodeUp) == 0,
		"expected len(nodeUp) == 0, got:", len(nodeUp))
}

func TestNodeHeathCacheHealthy(t *testing.T) {
	tmpfile := tests.Tempfile()
	defer os.Remove(tmpfile)

	// Create the app (I'm being lazy here. An app is not strictly
	// needed but it is convenient.
	app := NewTestApp(tmpfile)
	defer app.Close()

	err := setupSampleDbWithTopology(app,
		2,    // clusters
		3,    // nodes_per_cluster
		4,    // devices_per_node,
		6*TB, // disksize)
	)
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	hc := NewNodeHealthCache(app.db, app.executor)
	tests.Assert(t, hc != nil, "expected hc != nil, got:", hc)

	err = hc.Refresh()
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	nodeUp := hc.Status()
	tests.Assert(t, len(nodeUp) == 6,
		"expected len(nodeUp) == 6, got:", len(nodeUp))
	for _, v := range nodeUp {
		tests.Assert(t, v)
	}
}

func TestNodeHeathCacheMonitor(t *testing.T) {
	tmpfile := tests.Tempfile()
	defer os.Remove(tmpfile)

	// Create the app (I'm being lazy here. An app is not strictly
	// needed but it is convenient.
	app := NewTestApp(tmpfile)
	defer app.Close()

	err := setupSampleDbWithTopology(app,
		2,    // clusters
		3,    // nodes_per_cluster
		4,    // devices_per_node,
		6*TB, // disksize)
	)
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	cc := 0
	app.xo.MockGlusterdCheck = func(host string) error {
		cc++
		return nil
	}

	hc := NewNodeHealthCache(app.db, app.executor)
	tests.Assert(t, hc != nil, "expected hc != nil, got:", hc)

	hc.CheckInterval = time.Millisecond * 10
	hc.Monitor()

	time.Sleep(time.Millisecond * 60)
	hc.Stop()

	tests.Assert(t, cc >= (2*6), "expected cc >= (2 * 6), got:", cc)
}

func TestNodeHeathCacheSomeUnhealthy(t *testing.T) {
	tmpfile := tests.Tempfile()
	defer os.Remove(tmpfile)

	// Create the app (I'm being lazy here. An app is not strictly
	// needed but it is convenient.
	app := NewTestApp(tmpfile)
	defer app.Close()

	err := setupSampleDbWithTopology(app,
		2,    // clusters
		3,    // nodes_per_cluster
		4,    // devices_per_node,
		6*TB, // disksize)
	)
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	cc := 0
	app.xo.MockGlusterdCheck = func(host string) error {
		var e error
		if cc&1 == 1 {
			e = fmt.Errorf("Bloop %v", cc)
		}
		cc++
		return e
	}
	hc := NewNodeHealthCache(app.db, app.executor)
	tests.Assert(t, hc != nil, "expected hc != nil, got:", hc)

	err = hc.Refresh()
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	nodeUp := hc.Status()
	tests.Assert(t, len(nodeUp) == 6,
		"expected len(nodeUp) == 6, got:", len(nodeUp))
	var up, down int
	for _, v := range nodeUp {
		if v {
			up++
		} else {
			down++
		}
	}
	tests.Assert(t, up == 3, "expected len(up) == 3, got:", up)
	tests.Assert(t, down == 3, "expected len(down) == 3, got:", down)
}

func TestNodeHeathCacheMultiRefresh(t *testing.T) {
	tmpfile := tests.Tempfile()
	defer os.Remove(tmpfile)

	// Create the app (I'm being lazy here. An app is not strictly
	// needed but it is convenient.
	app := NewTestApp(tmpfile)
	defer app.Close()

	err := setupSampleDbWithTopology(app,
		2,    // clusters
		3,    // nodes_per_cluster
		4,    // devices_per_node,
		6*TB, // disksize)
	)
	tests.Assert(t, err == nil, "expected err == nil, got:", err)

	cc := 0
	app.xo.MockGlusterdCheck = func(host string) error {
		var e error
		if cc&1 == 1 {
			e = fmt.Errorf("Bloop %v", cc)
		}
		cc++
		return e
	}

	for i := 0; i < 5; i++ {
		cc = 0
		hc := NewNodeHealthCache(app.db, app.executor)
		tests.Assert(t, hc != nil, "expected hc != nil, got:", hc)

		err = hc.Refresh()
		tests.Assert(t, err == nil, "expected err == nil, got:", err)

		nodeUp := hc.Status()
		tests.Assert(t, len(nodeUp) == 6,
			"expected len(nodeUp) == 6, got:", len(nodeUp))
		var up, down int
		for _, v := range nodeUp {
			if v {
				up++
			} else {
				down++
			}
		}
		tests.Assert(t, up == 3, "expected len(up) == 3, got:", up)
		tests.Assert(t, down == 3, "expected len(down) == 3, got:", down)
	}
}
