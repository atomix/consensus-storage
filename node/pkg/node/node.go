// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package node

import (
	counterv1 "github.com/atomix/runtime/primitives/pkg/counter/v1"
	countermapv1 "github.com/atomix/runtime/primitives/pkg/countermap/v1"
	electionv1 "github.com/atomix/runtime/primitives/pkg/election/v1"
	indexedmapv1 "github.com/atomix/runtime/primitives/pkg/indexedmap/v1"
	lockv1 "github.com/atomix/runtime/primitives/pkg/lock/v1"
	mapv1 "github.com/atomix/runtime/primitives/pkg/map/v1"
	multimapv1 "github.com/atomix/runtime/primitives/pkg/multimap/v1"
	setv1 "github.com/atomix/runtime/primitives/pkg/set/v1"
	valuev1 "github.com/atomix/runtime/primitives/pkg/value/v1"
	"github.com/atomix/runtime/sdk/pkg/network"
	"github.com/atomix/runtime/sdk/pkg/protocol/node"
	"github.com/atomix/runtime/sdk/pkg/protocol/statemachine"
	"google.golang.org/grpc"
)

func New(network network.Network, config NodeConfig, opts ...node.Option) *node.Node {
	registry := statemachine.NewPrimitiveTypeRegistry()
	counterv1.RegisterStateMachine(registry)
	countermapv1.RegisterStateMachine(registry)
	electionv1.RegisterStateMachine(registry)
	indexedmapv1.RegisterStateMachine(registry)
	lockv1.RegisterStateMachine(registry)
	mapv1.RegisterStateMachine(registry)
	multimapv1.RegisterStateMachine(registry)
	setv1.RegisterStateMachine(registry)
	valuev1.RegisterStateMachine(registry)
	protocol := newProtocol(config, registry)
	node := node.NewNode(network, protocol, opts...)
	counterv1.RegisterServer(node)
	countermapv1.RegisterServer(node)
	electionv1.RegisterServer(node)
	indexedmapv1.RegisterServer(node)
	lockv1.RegisterServer(node)
	mapv1.RegisterServer(node)
	multimapv1.RegisterServer(node)
	setv1.RegisterServer(node)
	valuev1.RegisterServer(node)
	node.RegisterService(func(server *grpc.Server) {
		RegisterNodeServer(server, newNodeServer(protocol))
	})
	return node
}
