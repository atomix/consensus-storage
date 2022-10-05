// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cmd := &cobra.Command{
		Use: "atomix-multi-raft-node",
		Run: func(cmd *cobra.Command, args []string) {
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
			node := node.New(
				runtime.NewNetwork(),
				node.WithHost(apiHost),
				node.WithPort(apiPort),
				node.WithConfig(multiraftv1.NodeConfig{
					Host:            raftHost,
					Port:            int32(raftPort),
					MultiRaftConfig: config,
				}))

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
