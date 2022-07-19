// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"fmt"
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft/node/pkg/primitive"
	"github.com/atomix/runtime/pkg/logging"
	"github.com/atomix/runtime/pkg/runtime"
	"google.golang.org/grpc"
	"os"
)

var log = logging.GetLogger()

func New(network runtime.Network, opts ...Option) *MultiRaftNode {
	var options Options
	options.apply(opts...)
	registry := primitive.NewRegistry()
	for _, primitiveType := range options.PrimitiveTypes {
		registry.Register(primitiveType)
	}
	return &MultiRaftNode{
		Options:  options,
		network:  network,
		protocol: newProtocol(registry, options),
		server:   grpc.NewServer(),
	}
}

type MultiRaftNode struct {
	Options
	network  runtime.Network
	protocol *Protocol
	server   *grpc.Server
}

func (s *MultiRaftNode) Start() error {
	address := fmt.Sprintf("%s:%d", s.PrimitiveService.Host, s.PrimitiveService.Port)
	lis, err := s.network.Listen(address)
	if err != nil {
		return err
	}

	multiraftv1.RegisterPartitionServer(s.server, newPartitionServer(s.protocol))
	multiraftv1.RegisterSessionServer(s.server, newSessionServer(s.protocol))
	multiraftv1.RegisterNodeServer(s.server, newNodeServer(s.protocol.node))
	for _, primitiveType := range s.PrimitiveTypes {
		primitiveType.RegisterServices(s.server, s.protocol)
	}
	go func() {
		if err := s.server.Serve(lis); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	return nil
}

func (s *MultiRaftNode) Stop() error {
	s.server.Stop()
	return nil
}
