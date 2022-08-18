// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"fmt"
	atomiccounterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/counter/v1"
	atomicmapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/map/v1"
	atomicvaluev1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/value/v1"
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	mapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node/server"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	atomiccounterv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/atomic/counter/v1"
	atomicmapv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/atomic/map/v1"
	atomicvaluev1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/atomic/value/v1"
	counterv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/counter/v1"
	mapv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/map/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	atomiccountersmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/atomic/counter/v1"
	atomicmapsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/atomic/map/v1"
	atomicvaluesmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/atomic/value/v1"
	countersmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/counter/v1"
	mapsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/map/v1"
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
	atomiccountersmv1.Register(registry)
	atomicmapsmv1.Register(registry)
	atomicvaluesmv1.Register(registry)
	countersmv1.Register(registry)
	mapsmv1.Register(registry)
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

	atomiccounterv1.RegisterAtomicCounterServer(s.server, atomiccounterv1server.NewAtomicCounterServer(s.protocol))
	atomicmapv1.RegisterAtomicMapServer(s.server, atomicmapv1server.NewAtomicMapServer(s.protocol))
	atomicvaluev1.RegisterAtomicValueServer(s.server, atomicvaluev1server.NewAtomicValueServer(s.protocol))
	counterv1.RegisterCounterServer(s.server, counterv1server.NewCounterServer(s.protocol))
	mapv1.RegisterMapServer(s.server, mapv1server.NewMapServer(s.protocol))

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
