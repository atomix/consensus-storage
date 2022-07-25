// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitives

import (
	"github.com/atomix/multi-raft-storage/node/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"google.golang.org/grpc"
)

var log = logging.GetLogger()

func RegisterPrimitiveServers(server *grpc.Server, node *protocol.Node) {
	RegisterCounterServer(server, node)
	RegisterMapServer(server, node)
}
