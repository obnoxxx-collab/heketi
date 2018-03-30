//
// Copyright (c) 2015 The heketi Authors
//
// This file is licensed to you under your choice of the GNU Lesser
// General Public License, version 3 or any later version (LGPLv3 or
// later), or the GNU General Public License, version 2 (GPLv2), in all
// cases as published by the Free Software Foundation.
//

package glusterfs

import (
	"github.com/heketi/heketi/executors"
	wdb "github.com/heketi/heketi/pkg/db"
	"github.com/heketi/heketi/pkg/utils"
)

func CreateBricks(db wdb.RODB, executor executors.Executor, brick_entries []*BrickEntry) error {
	sg := utils.NewStatusGroup()

	// Create a goroutine for each brick
	for _, brick := range brick_entries {
		sg.Add(1)
		go func(b *BrickEntry) {
			defer sg.Done()
			sg.Err(b.Create(db, executor))
		}(brick)
	}

	// Wait here until all goroutines have returned.  If
	// any of errored, it would be cought here
	err := sg.Result()
	if err != nil {
		logger.Err(err)

		// Destroy all bricks and cleanup
		DestroyBricks(db, executor, brick_entries)
	}

	return err
}

func DestroyBricks(db wdb.RODB, executor executors.Executor, brick_entries []*BrickEntry) (map[string]bool, error) {
	sg := utils.NewStatusGroup()

	// return a map with the deviceId as key, and a bool if the space has been free'd
	reclaimed := map[string]bool{}

	// Create a goroutine for each brick
	for _, brick := range brick_entries {
		sg.Add(1)
		go func(b *BrickEntry, f map[string]bool) {
			defer sg.Done()
			spaceReclaimed, err := b.Destroy(db, executor)
			if err == nil {
				// mark space from device as freed
				f[b.Info.DeviceId] = spaceReclaimed
			}
			sg.Err(err)
		}(brick, reclaimed)
	}

	// Wait here until all goroutines have returned.  If
	// any of errored, it would be cought here
	err := sg.Result()
	if err != nil {
		logger.Err(err)
	}

	return reclaimed, err
}
