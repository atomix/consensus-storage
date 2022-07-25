// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	"github.com/atomix/multi-raft-storage/driver/pkg/primitives"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

const (
	name    = "MultiRaft"
	version = "v1beta1"
)

func New(network runtime.Network) runtime.Driver {
	return runtime.NewDriver[multiraftv1.ClusterConfig](name, version, func(ctx context.Context, config multiraftv1.ClusterConfig) (runtime.Client, error) {
		client := client.NewClient(network)
		if err := client.Connect(ctx, config); err != nil {
			return nil, err
		}
		return newClient(client), nil
	})
}

func newClient(client *client.Client) runtime.Client {
	return &driverClient{
		client: client,
	}
}

type driverClient struct {
	client *client.Client
}

func (c *driverClient) Counter() counterv1.CounterServer {
	return primitives.NewCounterServer(c.client.Protocol)
}

func (c *driverClient) Map() mapv1.MapServer {
	return primitives.NewMapServer(c.client.Protocol)
}

func (c *driverClient) Configure(ctx context.Context, config multiraftv1.ClusterConfig) error {
	return c.client.Configure(ctx, config)
}

func (c *driverClient) Close(ctx context.Context) error {
	return c.client.Close(ctx)
}
