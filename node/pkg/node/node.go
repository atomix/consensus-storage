// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"fmt"
	counterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	countermapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/countermap/v1"
	electionv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/election/v1"
	indexedmapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/indexedmap/v1"
	lockv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/lock/v1"
	mapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	multimapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/multimap/v1"
	setv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/set/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	valuev1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/value/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/node/server"
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	counterv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/counter/v1"
	countermapv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/countermap/v1"
	electionv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/election/v1"
	indexedmapv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/indexedmap/v1"
	lockv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/lock/v1"
	mapv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/map/v1"
	multimapv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/multimap/v1"
	setv1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/set/v1"
	valuev1server "github.com/atomix/multi-raft-storage/node/pkg/protocol/value/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive"
	countersmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/counter/v1"
	countermapsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/countermap/v1"
	electionsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/election/v1"
	indexedmapsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/indexedmap/v1"
	locksmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/lock/v1"
	mapsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/map/v1"
	multimapsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/multimap/v1"
	setsmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/set/v1"
	valuesmv1 "github.com/atomix/multi-raft-storage/node/pkg/statemachine/primitive/value/v1"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/atomix/runtime/sdk/pkg/network"
	"google.golang.org/grpc"
	"os"
)

var log = logging.GetLogger()

func New(network network.Network, opts ...Option) *MultiRaftNode {
	var options Options
	options.apply(opts...)
	registry := primitive.NewTypeRegistry()
	countersmv1.Register(registry)
	countermapsmv1.Register(registry)
	indexedmapsmv1.Register(registry)
	locksmv1.Register(registry)
	electionsmv1.Register(registry)
	mapsmv1.Register(registry)
	multimapsmv1.Register(registry)
	setsmv1.Register(registry)
	valuesmv1.Register(registry)
	return &MultiRaftNode{
		Options:  options,
		network:  network,
		protocol: protocol.NewNode(&options.Config, registry),
		server:   grpc.NewServer(),
	}
}

type MultiRaftNode struct {
	Options
	network  network.Network
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

	counterv1.RegisterCounterServer(s.server, counterv1server.NewCounterServer(s.protocol))
	countermapv1.RegisterCounterMapServer(s.server, countermapv1server.NewCounterMapServer(s.protocol))
	indexedmapv1.RegisterIndexedMapServer(s.server, indexedmapv1server.NewIndexedMapServer(s.protocol))
	electionv1.RegisterLeaderElectionServer(s.server, electionv1server.NewLeaderElectionServer(s.protocol))
	lockv1.RegisterLockServer(s.server, lockv1server.NewLockServer(s.protocol))
	mapv1.RegisterMapServer(s.server, mapv1server.NewMapServer(s.protocol))
	multimapv1.RegisterMultiMapServer(s.server, multimapv1server.NewMultiMapServer(s.protocol))
	setv1.RegisterSetServer(s.server, setv1server.NewSetServer(s.protocol))
	valuev1.RegisterValueServer(s.server, valuev1server.NewValueServer(s.protocol))

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
