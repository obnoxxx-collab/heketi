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
	"github.com/boltdb/bolt"

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

func DestroyBricks(db wdb.RODB, executor executors.Executor, brick_entries []*BrickEntry) error {
	sg := utils.NewStatusGroup()

	var device_entries []*DeviceEntry

	// Create a goroutine for each brick
	for _, brick := range brick_entries {
		// TODO: Find the DeviceEntry where the space is free'd on brick.Destroy()
		var device *DeviceEntry
		err := db.View(func(tx *bolt.Tx) error {
			var err error
			device, err = NewDeviceEntryFromId(tx, brick.Info.DeviceId)
			return err
		})
		if err != nil {
			return err
		}
		// we'll modify the device in the go routing, it needs Save()'ing later
		device_entries = append(device_entries, device)

		sg.Add(1)
		go func(b *BrickEntry, d *DeviceEntry) {
			defer sg.Done()
			sizeFreed, err := b.Destroy(db, executor)
			if err == nil {
				// mark space from device as freed
				device.StorageFree(sizeFreed)
			}
			sg.Err(err)
		}(brick, device)
	}

	// Wait here until all goroutines have returned.  If
	// any of errored, it would be cought here
	err := sg.Result()
	if err != nil {
		logger.Err(err)
	}

	// TODO: this is probably the wrong location to save the DeviceEntry
	for _, d := range device_entries {
		// TODO: db.Update() is not possible here with a wdb.RODB
		logger.Debug("TODO: storare device %v with %v free storage", d.Id, d.Info.Storage.Free)
//		err := db.Update(func(tx *bolt.Tx) error {
//			d.Save(tx)
//			return nil
//		})
//		if err != nil {
//			return err
//		}
	}

	return err
}
