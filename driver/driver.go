// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"github.com/atomix/runtime/sdk/pkg/network"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

var driverID = runtime.DriverID{
	Name:    "MultiRaft",
	Version: "v2beta1",
}

func New(network network.Network) runtime.Driver {
	return &multiRaftDriver{
		network: network,
	}
}

type multiRaftDriver struct {
	network network.Network
}

func (d *multiRaftDriver) ID() runtime.DriverID {
	return driverID
}

func (d *multiRaftDriver) Connect(ctx context.Context, spec runtime.ConnSpec) (runtime.Conn, error) {
	conn := newConn(d.network)
	if err := conn.Connect(ctx, spec); err != nil {
		return nil, err
	}
	return conn, nil
}

func (d *multiRaftDriver) String() string {
	return d.ID().String()
}
