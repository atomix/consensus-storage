// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	counterv1api "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	counterv1 "github.com/atomix/runtime/primitives/pkg/counter/v1"
	"github.com/atomix/runtime/sdk/pkg/network"
	"github.com/atomix/runtime/sdk/pkg/protocol"
	"github.com/atomix/runtime/sdk/pkg/protocol/client"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

func newConn(network network.Network) *multiRaftConn {
	return &multiRaftConn{
		ProtocolClient: client.NewClient(network),
	}
}

type multiRaftConn struct {
	*client.ProtocolClient
}

func (c *multiRaftConn) Connect(ctx context.Context, spec runtime.ConnSpec) error {
	var config protocol.ProtocolConfig
	if err := spec.UnmarshalConfig(&config); err != nil {
		return err
	}
	return c.ProtocolClient.Connect(ctx, config)
}

func (c *multiRaftConn) Configure(ctx context.Context, spec runtime.ConnSpec) error {
	var config protocol.ProtocolConfig
	if err := spec.UnmarshalConfig(&config); err != nil {
		return err
	}
	return c.ProtocolClient.Configure(ctx, config)
}

func (c *multiRaftConn) NewCounter(spec runtime.PrimitiveSpec) (counterv1api.CounterServer, error) {
	return counterv1.NewCounterProxy(c.Protocol, spec)
}
