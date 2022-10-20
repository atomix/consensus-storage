// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	counterv1 "github.com/atomix/runtime/primitives/pkg/counter/v1"
	"github.com/atomix/runtime/sdk/pkg/network"
	"github.com/atomix/runtime/sdk/pkg/protocol/node"
	"github.com/atomix/runtime/sdk/pkg/protocol/statemachine"
	"google.golang.org/grpc"
)

func New(network network.Network, config NodeConfig, opts ...node.Option) *node.Node {
	registry := statemachine.NewPrimitiveTypeRegistry()
	counterv1.RegisterStateMachine(registry)
	protocol := newProtocol(config, registry)
	node := node.NewNode(network, protocol, opts...)
	counterv1.RegisterServer(node)
	node.RegisterService(func(server *grpc.Server) {
		RegisterNodeServer(server, newNodeServer(protocol))
	})
	return node
}
