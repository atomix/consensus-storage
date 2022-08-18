// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	multiraftatomiccounterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/counter/v1"
	multiraftatomiccountermapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/countermap/v1"
	multiraftatomicmapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/atomic/map/v1"
	multiraftcounterv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/counter/v1"
	multiraftmapv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/map/v1"
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/driver/pkg/client"
	atomiccounterv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/atomic/counter/v1"
	atomiccountermapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/atomic/countermap/v1"
	atomicmapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/atomic/map/v1"
	counterv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/counter/v1"
	mapv1server "github.com/atomix/multi-raft-storage/driver/pkg/primitives/map/v1"
	atomiccounterv1 "github.com/atomix/runtime/api/atomix/runtime/atomic/counter/v1"
	atomiccountermapv1 "github.com/atomix/runtime/api/atomix/runtime/atomic/countermap/v1"
	atomicmapv1 "github.com/atomix/runtime/api/atomix/runtime/atomic/map/v1"
	counterv1 "github.com/atomix/runtime/api/atomix/runtime/counter/v1"
	mapv1 "github.com/atomix/runtime/api/atomix/runtime/map/v1"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

const (
	name    = "MultiRaft"
	version = "v1beta1"
)

func New(network runtime.Network) runtime.Driver {
	return runtime.NewDriver[multiraftv1.DriverConfig](name, version, func(ctx context.Context, config multiraftv1.DriverConfig) (runtime.Conn, error) {
		client := client.NewClient(network)
		if err := client.Connect(ctx, config); err != nil {
			return nil, err
		}
		return runtime.NewConn[multiraftv1.DriverConfig](client,
			runtime.WithAtomicCounterFactory[multiraftatomiccounterv1.AtomicCounterConfig](func(config multiraftatomiccounterv1.AtomicCounterConfig) (atomiccounterv1.AtomicCounterServer, error) {
				return atomiccounterv1server.NewAtomicCounterServer(client.Protocol, config), nil
			}),
			runtime.WithAtomicCounterMapFactory[multiraftatomiccountermapv1.AtomicCounterMapConfig](func(config multiraftatomiccountermapv1.AtomicCounterMapConfig) (atomiccountermapv1.AtomicCounterMapServer, error) {
				return atomiccountermapv1server.NewAtomicCounterMapServer(client.Protocol), nil
			}),
			runtime.WithAtomicMapFactory[multiraftatomicmapv1.AtomicMapConfig](func(config multiraftatomicmapv1.AtomicMapConfig) (atomicmapv1.AtomicMapServer, error) {
				return atomicmapv1server.NewAtomicMapServer(client.Protocol, config), nil
			}),
			runtime.WithCounterFactory[multiraftcounterv1.CounterConfig](func(config multiraftcounterv1.CounterConfig) (counterv1.CounterServer, error) {
				return counterv1server.NewCounterServer(client.Protocol, config), nil
			}),
			runtime.WithMapFactory[multiraftmapv1.MapConfig](func(config multiraftmapv1.MapConfig) (mapv1.MapServer, error) {
				return mapv1server.NewMapServer(client.Protocol, config), nil
			})), nil
	})
}
