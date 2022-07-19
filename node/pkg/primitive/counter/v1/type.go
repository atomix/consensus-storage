// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	counterv1 "github.com/atomix/multi-raft/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft/node/pkg/primitive"
	"google.golang.org/grpc"
)

const Name = "Counter"
const APIVersion = "v1"

var Type = primitive.NewType(Name, APIVersion, registerServer, newStateMachine)

func registerServer(server *grpc.Server, protocol primitive.Protocol) {
	counterv1.RegisterCounterServer(server, newServer(protocol))
}

func newStateMachine(ctx primitive.Context) primitive.StateMachine {

}
