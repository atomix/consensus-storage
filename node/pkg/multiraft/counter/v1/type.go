// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	counterv1 "github.com/atomix/multi-raft/api/atomix/multiraft/counter/v1"
	"github.com/atomix/multi-raft/node/pkg/multiraft"
	"google.golang.org/grpc"
)

var Type = multiraft.NewType(register)

func register(server *grpc.Server, protocol *multiraft.Protocol) {
	counterv1.RegisterCounterServer(server, newServer(protocol))
}

func create() multiraft.StateMachine {

}
