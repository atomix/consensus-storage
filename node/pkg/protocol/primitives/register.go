// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitives

import (
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"google.golang.org/grpc"
)

func RegisterPrimitiveServers(server *grpc.Server, node *protocol.Node) {
	RegisterCounterServer(server, node)
}
