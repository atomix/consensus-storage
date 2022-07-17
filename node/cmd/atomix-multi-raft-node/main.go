// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/node/pkg/multiraft"
	"github.com/atomix/runtime/pkg/logging"
	counterv1 "github.com/atomix/runtime/primitives/counter/v1"
	electionv1 "github.com/atomix/runtime/primitives/election/v1"
	indexedmapv1 "github.com/atomix/runtime/primitives/indexed_map/v1"
	listv1 "github.com/atomix/runtime/primitives/list/v1"
	lockv1 "github.com/atomix/runtime/primitives/lock/v1"
	mapv1 "github.com/atomix/runtime/primitives/map/v1"
	setv1 "github.com/atomix/runtime/primitives/set/v1"
	topicv1 "github.com/atomix/runtime/primitives/topic/v1"
	valuev1 "github.com/atomix/runtime/primitives/value/v1"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logging.SetLevel(logging.DebugLevel)

	cmd := &cobra.Command{
		Use: "atomix-multi-raft-node",
		Run: func(cmd *cobra.Command, args []string) {
			nodeID, err := cmd.Flags().GetString("node")
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err.Error())
				os.Exit(1)
			}
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err.Error())
				os.Exit(1)
			}
			apiHost, err := cmd.Flags().GetString("api-host")
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err.Error())
				os.Exit(1)
			}
			apiPort, err := cmd.Flags().GetInt("api-port")
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err.Error())
				os.Exit(1)
			}
			raftHost, err := cmd.Flags().GetString("raft-host")
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err.Error())
				os.Exit(1)
			}
			raftPort, err := cmd.Flags().GetInt("raft-port")
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err.Error())
				os.Exit(1)
			}

			config := multiraftv1.MultiRaftConfig{}
			configBytes, err := ioutil.ReadFile(configPath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if err := jsonpb.Unmarshal(bytes.NewReader(configBytes), &config); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Create the multi-raft node
			node := multiraft.NewNode(
				multiraft.NewNetwork(),
				multiraft.WithNodeID(nodeID),
				multiraft.WithConfig(config),
				multiraft.WithAPIHost(apiHost),
				multiraft.WithAPIPort(apiPort),
				multiraft.WithRaftHost(raftHost),
				multiraft.WithRaftPort(raftPort),
				multiraft.WithPrimitiveTypes(
					counterv1.Type,
					electionv1.Type,
					indexedmapv1.Type,
					listv1.Type,
					lockv1.Type,
					mapv1.Type,
					setv1.Type,
					topicv1.Type,
					valuev1.Type))

			// Start the node
			if err := node.Start(); err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err.Error())
				os.Exit(1)
			}

			// Wait for an interrupt signal
			ch := make(chan os.Signal, 2)
			signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
			<-ch

			// Stop the node
			if err := node.Stop(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringP("node", "n", "", "the ID of this node")
	cmd.Flags().StringP("config", "c", "", "the path to the multi-raft cluster configuration")
	cmd.Flags().String("api-host", "", "the host to which to bind the API server")
	cmd.Flags().Int("api-port", 8080, "the port to which to bind the API server")
	cmd.Flags().String("raft-host", "", "the host to which to bind the Multi-Raft server")
	cmd.Flags().Int("raft-port", 5000, "the port to which to bind the Multi-Raft server")

	_ = cmd.MarkFlagRequired("node")
	_ = cmd.MarkFlagRequired("config")
	_ = cmd.MarkFlagFilename("config")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
