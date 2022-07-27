// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node/server"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	servers "github.com/atomix/multi-raft-storage/node/pkg/protocol/primitives"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	statemachines "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitives"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/runtime"
	"google.golang.org/grpc"
	"os"
)

var log = logging.GetLogger()

func New(network runtime.Network, opts ...Option) *MultiRaftNode {
	var options Options
	options.apply(opts...)
	registry := statemachine.NewPrimitiveTypeRegistry()
	statemachines.RegisterPrimitiveTypes(registry)
	return &MultiRaftNode{
		Options:  options,
		network:  network,
		protocol: protocol.NewNode(&options.Config, registry),
		server:   grpc.NewServer(),
	}
}

type MultiRaftNode struct {
	Options
	network  runtime.Network
	protocol *protocol.Node
	server   *grpc.Server
}

func (s *MultiRaftNode) Start() error {
	log.Infow("Starting MultiRaftNode",
		logging.Stringer("Config", &s.Config))
	address := fmt.Sprintf("%s:%d", s.Host, s.Port)
	lis, err := s.network.Listen(address)
	if err != nil {
		log.Errorw("Error starting MultiRaftNode",
			logging.Stringer("Config", &s.Config),
			logging.Error("Error", err))
		return err
	}

	multiraftv1.RegisterNodeServer(s.server, server.NewNodeServer(s.protocol))
	multiraftv1.RegisterPartitionServer(s.server, server.NewPartitionServer(s.protocol))
	multiraftv1.RegisterSessionServer(s.server, server.NewSessionServer(s.protocol))
	servers.RegisterPrimitiveServers(s.server, s.protocol)
	go func() {
		if err := s.server.Serve(lis); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	return nil
}

func (s *MultiRaftNode) Stop() error {
	log.Infow("Stopping MultiRaftNode",
		logging.Stringer("Config", &s.Config))
	s.server.Stop()
	if err := s.protocol.Shutdown(); err != nil {
		log.Errorw("Error starting MultiRaftNode",
			logging.Stringer("Config", &s.Config),
			logging.Error("Error", err))
		return err
	}
	return nil
}
