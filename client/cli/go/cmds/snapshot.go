//
// Copyright (c) 2018 The heketi Authors
//
// This file is licensed to you under your choice of the GNU Lesser
// General Public License, version 3 or any later version (LGPLv3 or
// later), or the GNU General Public License, version 2 (GPLv2), in all
// cases as published by the Free Software Foundation.
//

package cmds

import (
	"encoding/json"
	"errors"
	"fmt"

	client "github.com/heketi/heketi/client/api/go-client"
	"github.com/heketi/heketi/pkg/glusterfs/api"
	"github.com/spf13/cobra"
)

var (
	clonename string
)

func init() {
	RootCmd.AddCommand(snapshotCommand)

	snapshotCommand.AddCommand(snapshotDeleteCommand)
	snapshotDeleteCommand.SilenceUsage = true

	snapshotCommand.AddCommand(snapshotCloneCommand)
	snapshotCloneCommand.Flags().StringVar(&clonename, "name", "",
		"\n\tOptional: Name of the newly cloned volume. Only set if really necessary")
	snapshotCloneCommand.SilenceUsage = true

	snapshotCommand.AddCommand(snapshotInfoCommand)
	snapshotInfoCommand.SilenceUsage = true

	snapshotCommand.AddCommand(snapshotListCommand)
	snapshotListCommand.SilenceUsage = true
}

var snapshotCommand = &cobra.Command{
	Use:   "snapshot",
	Short: "Heketi Snapshot Management",
	Long:  "Heketi Snapshot Management",
}

var snapshotDeleteCommand = &cobra.Command{
	Use:     "delete",
	Short:   "Deletes the snapshot",
	Long:    "Deletes the snapshot",
	Example: "  $ heketi-cli snapshot delete 886a86a868711bef83001",
	RunE: func(cmd *cobra.Command, args []string) error {
		// ensure proper number of args
		s := cmd.Flags().Args()
		if len(s) < 1 {
			return errors.New("Snapshot id missing")
		}

		snapshotId := cmd.Flags().Arg(0)

		heketi := client.NewClient(options.Url, options.User, options.Key)
		err := heketi.SnapshotDelete(snapshotId)
		if err == nil {
			fmt.Fprintf(stdout, "Snapshot %s deleted\n", snapshotId)
		}

		return err
	},
}

var snapshotCloneCommand = &cobra.Command{
	Use:     "clone",
	Short:   "Clones a snapshot into a new volume",
	Long:    "Clones a snapshot into a new volume",
	Example: "  $ heketi-cli snapshot clone 886a86a868711bef83001",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := cmd.Flags().Args()
		if len(s) < 1 {
			return errors.New("Snapshot id missing")
		}

		// Create request blob
		snapId := cmd.Flags().Arg(0)
		req := &api.SnapshotCloneRequest{}
		if clonename != "" {
			req.Name = clonename
		}

		heketi := client.NewClient(options.Url, options.User, options.Key)

		volume, err := heketi.SnapshotClone(snapId, req)
		if err != nil {
			return err
		}

		if options.Json {
			data, err := json.Marshal(volume)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "%v", volume)
		}

		return nil
	},
}

var snapshotInfoCommand = &cobra.Command{
	Use:     "info",
	Short:   "Shows the information of a snapshot",
	Long:    "Shows the information of a snapshot",
	Example: "  $ heketi-cli snapshot info 886a86a868711bef83001",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := cmd.Flags().Args()
		if len(s) < 1 {
			return errors.New("Snapshot id missing")
		}

		heketi := client.NewClient(options.Url, options.User, options.Key)

		snapshotId := cmd.Flags().Arg(0)
		snapshot, err := heketi.SnapshotInfo(snapshotId)
		if err != nil {
			return err
		}

		if options.Json {
			data, err := json.Marshal(snapshot)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, string(data))
		} else {
			fmt.Fprintf(stdout, "%v", snapshot)
		}
		return nil
	},
}

var snapshotListCommand = &cobra.Command{
	Use:     "list",
	Short:   "Lists all snapshots",
	Long:    "Lists all snapshots",
	Example: "  $ heketi-cli snapshot list",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a client
		heketi := client.NewClient(options.Url, options.User, options.Key)

		// List snapshots
		list, err := heketi.SnapshotList()
		if err != nil {
			return err
		}

		if options.Json {
			data, err := json.Marshal(list)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, string(data))
		} else {
			for _, id := range list.Snapshots {
				snap, err := heketi.SnapshotInfo(id)
				if err != nil {
					return err
				}

				fmt.Fprintf(stdout, "Id:%-35v Name:%v CreateTime:%v Type:%v\n",
					id,
					snap.Name,
					snap.CreateTime,
					snap.Type)
			}
		}

		return nil
	},
}
